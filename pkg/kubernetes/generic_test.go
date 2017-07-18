package kubernetes_test

import (
	"testing"

	//"github.com/Sirupsen/logrus"
	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"
	"github.com/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

func TestGeneric_Ensure(t *testing.T) {
	vault := vault_dev.New()
	if err := vault.Start(); err != nil {
		t.Skip("unable to initialise vault dev server for integration tests: ", err)
		return
	}
	defer vault.Stop()

	k, err := kubernetes.New(vault.Client(), "test-cluster")

	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	generic := k.NewGeneric()
	err = generic.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
		return
	}

	err = generic.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
		return
	}
}
