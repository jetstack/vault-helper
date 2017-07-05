package kubernetes_test

import (
	"testing"

	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/kubernetes"
)

func TestKubernetes_Ensure(t *testing.T) {
	k := kubernetes.New("test-cluster-2")

	err := k.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
	}
}
