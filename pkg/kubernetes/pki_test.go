package kubernetes

import (
	"testing"
	"time"

	//"github.com/Sirupsen/logrus"
	//"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/kubernetes"

	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

func TestPKI_Ensure(t *testing.T) {
	vault := vault_dev.New()
	if err := vault.Start(); err != nil {
		t.Skip("unable to initialise vault dev server for integration tests: ", err)
	}
	defer vault.Stop()

	k, err := New(vault.Client(), "test-cluster-inside")
	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	if exp, act := "test-cluster-inside/pki/etcd-k8s", k.etcdKubernetesPKI.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
	}
	if exp, act := "test-cluster-inside/pki/etcd-overlay", k.etcdOverlayPKI.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
	}
	if exp, act := "test-cluster-inside/pki/k8s", k.kubernetesPKI.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
	}
	if exp, act := "test-cluster-inside/generic", k.secretsGeneric.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
	}

	k.etcdKubernetesPKI.DefaultLeaseTTL = time.Hour * 0
	k.etcdOverlayPKI.DefaultLeaseTTL = time.Hour * 0
	k.kubernetesPKI.DefaultLeaseTTL = time.Hour * 0
	k.Ensure()
	if err != nil {
		t.Error("unexpected error", err)
		return
	}

}
