package kubernetes_test

import (
	"testing"

	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/kubernetes"
	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

func TestKubernetes_Ensure(t *testing.T) {
	vault := vault_dev.New()
	if err := vault.Start(); err != nil {
		t.Skip("unable to initialise vault dev server for integration tests: ", err)
	}
	defer vault.Stop()

	k, err := kubernetes.New(vault.Client(), "test-cluster18")

	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	err = k.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	err = k.GenerateSecretsMount()
	if err != nil {
		t.Error("unexpected error: ", err)
	}
}
