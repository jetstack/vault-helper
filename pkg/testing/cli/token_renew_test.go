package cli

import (
	"testing"
)

func TestRenewToken_Success(t *testing.T) {
	vault, err := InitVaultDev()
	if err != nil {
		t.Fatalf("failed to initiate vault for testing: %v", err)
	}
	defer vault.Stop()

	if _, err := InitKubernetes(vault); err != nil {
		t.Fatalf("failed to initiate kubernetes for testing: %v", err)
	}

	args := []string{"renew-token", "--init-role=test-master"}
	exitcode, err := RunCommand(args)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if exitcode != 0 {
		t.Errorf("unexpected error code, exp=0 got=%d", exitcode)
	}
}
