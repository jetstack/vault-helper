package vault_dev

import (
	"fmt"
	"net"
	"os/exec"
	"time"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

func getUnusedPort() int {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

type VaultDev struct {
	client *vault.Client
	server *exec.Cmd
}

func New() *VaultDev {
	return &VaultDev{}
}

func (v *VaultDev) Start() error {
	port := getUnusedPort()

	args := []string{
		"server",
		"-dev",
		"-dev-root-token-id=\"root-token\"",
		fmt.Sprintf("-dev-listen-address=\"127.0.0.1:%d\"", port),
	}

	logrus.Infof("starting vault: %#+v", args)

	v.server = exec.Command("vault", args...)

	err := v.server.Start()

	if err != nil {
		return err
	}

	v.client, err = vault.NewClient(&vault.Config{
		Address: fmt.Sprintf("http://127.0.0.1:%d", port),
	})
	if err != nil {
		return err
	}
	v.client.SetToken("root-token")

	tries := 10
	for {
		_, err := v.client.Auth().Token().LookupSelf()
		if err == nil {
			break
		}
		if tries <= 1 {
			return fmt.Errorf("vault dev server couldn't be started in time")
		}
		tries -= 1
		time.Sleep(time.Second)

	}

	return nil
}

func (v *VaultDev) Stop() {
	if err := v.server.Process.Kill(); err != nil {
		logrus.Warn("killing vault dev server failed: ", err)
	}

	if err := v.server.Wait(); err != nil {
		logrus.Warn("waiting for vault dev server to exit failed: ", err)
	}

}

func (v *VaultDev) Client() *vault.Client {
	return v.client
}
