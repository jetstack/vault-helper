package kubernetes_test

import (
	"testing"

	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/kubernetes"
)

func TestKubernetes_Ensure(t *testing.T) {
	k, err := kubernetes.New("test-cluster18")

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
