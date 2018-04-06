package cli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/vault-helper/pkg/kubernetes"
	"github.com/jetstack/vault-helper/pkg/testing/vault_dev"
)

const (
	binPath = "src/github.com/jetstack/vault-helper/vault-helper_linux_amd64"
)

var tmpDirs []string

func TestMain(m *testing.M) {

	vault, err := InitVaultDev()
	if err != nil {
		logrus.Fatalf("failed to initiate vault for testing: %v", err)
	}
	logrus.RegisterExitHandler(vault.Stop)
	defer vault.Stop()

	if _, err := InitKubernetes(vault); err != nil {
		logrus.Fatalf("failed to initiate kubernetes for testing: %v", err)
	}

	exitCode := m.Run()

	if err := CleanDirs(); err != nil {
		logrus.Errorf("error cleaning up tmp dirs: %v", err)
	}

	os.Exit(exitCode)
}

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
		Etcd:   "etcd",
		Master: "master",
		Worker: "worker",
		All:    "all",
	})

	if err := k.Ensure(); err != nil {
		return nil, fmt.Errorf("error ensuring kubernetes: %v", err)
	}

	return k, nil
}

func RunTest(args []string, expCode int, t *testing.T) {
	var stdout, stderr bytes.Buffer

	gotCode, err := runCommand(args, &stdout, &stderr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if expCode != gotCode {
		t.Errorf("unexpected error code, exp=%d got=%d", expCode, gotCode)
		fmt.Printf("%s\n", stdout.String())
		fmt.Printf("%s\n", stderr.String())
	}
}

func runCommand(args []string, stdout, stderr *bytes.Buffer) (int, error) {
	dir, err := initTokensDir()
	if err != nil {
		return -1, fmt.Errorf("failed to create tokens directory: %v", err)
	}

	args = append(args, fmt.Sprintf("--config-path=%s", dir))
	cmd := exec.Command(fmt.Sprintf("%s/%s", os.Getenv("GOPATH"), binPath), args...)

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	logrus.Infof("running command: [vault-helper %s]", strings.Join(args, " "))

	if err := cmd.Start(); err != nil {
		return -1, fmt.Errorf("error starting command: %v", err)
	}

	if err := cmd.Wait(); err != nil {

		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), nil
			}

			return -1, fmt.Errorf("failed to get command status: %v", err)

		} else {
			return -1, fmt.Errorf("error during wait for command: %v", err)
		}
	}

	return 0, nil
}

func CleanDirs() error {
	var result *multierror.Error

	for _, dir := range tmpDirs {
		if err := os.RemoveAll(dir); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func TmpDir() (string, error) {
	dir, err := ioutil.TempDir("", "test-cluster")
	if err != nil {
		return dir, fmt.Errorf("failed to create token directory: %v", err)
	}
	tmpDirs = append(tmpDirs, dir)

	return dir, nil
}

func initTokensDir() (string, error) {
	dir, err := TmpDir()
	if err != nil {
		return dir, err
	}

	initTokenFile := fmt.Sprintf("%s/init-token", dir)
	tokenFile := fmt.Sprintf("%s/token", dir)

	if err := ioutil.WriteFile(initTokenFile, []byte("root-token-dev"), 0644); err != nil {
		return dir, fmt.Errorf("failed to write root-token-dev token to file: %v", err)
	}

	f, err := os.Create(tokenFile)
	if err != nil {
		return dir, fmt.Errorf("failed to create token file: %v", err)
	}
	if err := f.Close(); err != nil {
		return dir, fmt.Errorf("failed to close token file: %v", err)
	}

	return dir, nil
}
