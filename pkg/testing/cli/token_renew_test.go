package cli

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestHelloWorld(t *testing.T) {
	vault, err := InitVaultDev()
	if err != nil {
		t.Fatalf("failed to initiate vault for testing: %v", err)
	}
	defer vault.Stop()
	logrus.RegisterExitHandler(vault.Stop)

	if _, err := InitKubernetes(vault); err != nil {
		t.Fatalf("failed to initiate kubernetes for testing: %v", err)
	}

	args := []string{"renew-token", "--init-role=test-master"}
	if err := RunCommand(args); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
