package kubernetes_test

import (
	"testing"

	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/kubernetes"
)

func TestKubernetes_Ensure(t *testing.T) {
	k := kubernetes.New("test-cluster14")

	if k == nil {
		t.Error("No Cluster!")
		return
	}

	err := k.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
	}
}
