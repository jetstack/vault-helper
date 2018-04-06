package cli

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

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

	returnCode := m.Run()

	if err := CleanDirs(); err != nil {
		logrus.Errorf("error removing temp dirs: %v", err)
	}

	os.Exit(returnCode)
}

func TestRenewToken_Success(t *testing.T) {

	args := [][]string{
		[]string{"renew-token", "--init-role=test-master"},
		[]string{"renew-token", "--init-role=test-worker"},
		[]string{"renew-token", "--init-role=test-etcd"},
		[]string{"renew-token", "--init-role=test-all"},
	}

	for _, arg := range args {
		RunTest(arg, 0, t)
	}
}

func TestRenewToken_Fail(t *testing.T) {

	args := [][]string{
		[]string{"renew-token", "--init-role=test-foo"},
		[]string{"renew-token", "--init-role=foo"},
		[]string{"renew-token", "--init-role="},
	}
	for _, arg := range args {
		RunTest(arg, 1, t)
	}
}
