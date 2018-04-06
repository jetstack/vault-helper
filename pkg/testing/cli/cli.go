package cli

import (
	"fmt"
	"strings"
	"os"
	"io/ioutil"

	"github.com/sirupsen/logrus"

	"github.com/jetstack/vault-helper/cmd"
	"github.com/jetstack/vault-helper/pkg/kubernetes"
	"github.com/jetstack/vault-helper/pkg/testing/vault_dev"
)

func InitVaultDev() (*vault_dev.VaultDev, error) {
	vaultDev := vault_dev.New()

	if err := vaultDev.Start(); err != nil {
		return nil, fmt.Errorf("unable to initialise vault dev server for testing: %v", err)
	}

	addr := fmt.Sprintf("http://127.0.0.1:%d", vaultDev.Port())

	if err := os.Setenv("VAULT_ADDR", addr); err != nil {
		vaultDev.Stop()
		return nil, fmt.Errorf("failed to set vault address environment variable: %v", err)
	}

	if err := os.Setenv("VAULT_TOKEN", "root-token-dev"); err != nil {
		vaultDev.Stop()
		return nil, fmt.Errorf("failed to set vault root token environment variable: %v", err)
	}

	return vaultDev, nil
}


func InitKubernetes(vaultDev *vault_dev.VaultDev) (*kubernetes.Kubernetes, error) {
	k := kubernetes.New(vaultDev.Client(), logrus.NewEntry(logrus.New()))
	k.SetClusterID("test")
	k.SetInitFlags(kubernetes.FlagInitTokens{
		Etcd: "etcd",
		Master: "master",
		Worker: "worker",
		All: "all",
	})

	if err := k.Ensure(); err != nil {
		return nil, fmt.Errorf("error ensuring kubernetes: %v", err)
	}

	return k, nil
}

func RunCommand(args []string) error {
	dir, err := initTokensDir()
	if err != nil {
		return fmt.Errorf("failed to create tokens directory: %v", err)
	}

	args = append(args, fmt.Sprintf("--config-path=%s", dir))

	command := cmd.RootCmd
	command.SetArgs(args)
	command.PersistentFlags().Int("log-level", 1, "Set the log level of output. 0-Fatal 1-Info 2-Debug")

	command.Flag("log-level").Shorthand = "l"
	logrus.Infof("running command: [%s %s]", command.Name(), strings.Join(args, " "))

	if err := command.Execute(); err != nil {
		return fmt.Errorf("error running command: %v", err)
	}

	return nil
}

func initTokensDir() (string, error) {
	dir, err := ioutil.TempDir("", "test-cluster-dir")
	if err != nil {
		return "", fmt.Errorf("failed to create token directory: %v", err)
	}

	initTokenFile := fmt.Sprintf("%s/init-token", dir)
	tokenFile := fmt.Sprintf("%s/token", dir)

	if err := ioutil.WriteFile(initTokenFile, []byte("root-token-dev"), 0644); err != nil {
		return "", fmt.Errorf("failed to write root-token-dev token to file: %v", err)
	}

	f, err := os.Create(tokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to create token file: %v", err)
	}
	f.Close()

	return dir, nil
}
