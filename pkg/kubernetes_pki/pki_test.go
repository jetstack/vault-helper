package kubernetes_pki_test

import (
	"testing"

	vault "github.com/hashicorp/vault/api"

	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/kubernetes_pki"
)

func TestPKI_Ensure(t *testing.T) {

	testPath := "test-vault-helper-pki-ensure"

	vaultClient, err := vault.NewClient(nil)
	if err != nil {
		t.Skip("Unable to create vault client, skipping integration tests: ", err)
	}

	// should create non existing mount
	pki := kubernetes_pki.NewPKI(vaultClient, testPath)
	pki.Description = "first description"
	if err = pki.Ensure(); err != nil {
		t.Error("Unexpected error:", err)
	}

	// should update description
	secondDescription := "second description"
	pki.Description = secondDescription
	if err = pki.Ensure(); err != nil {
		t.Error("Unexpected error:", err)
	}
	mount, err := kubernetes_pki.GetMountByPath(vaultClient, testPath)
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	if exp, act := secondDescription, mount.Description; exp != act {
		t.Errorf("Did not update description: exp=%s act=%s", exp, act)
	}
}
