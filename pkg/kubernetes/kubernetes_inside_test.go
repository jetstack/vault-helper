package kubernetes

import (
	"testing"
)

func TestKubernetes_Backend_Path(t *testing.T) {
	k := New("test-cluster")
	if k == nil {
		t.Error("No Cluster!")
		return
	}

	if exp, act := "test-cluster/pki/etcd-k8s", k.etcdKubernetesPKI.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
	}
	if exp, act := "test-cluster/pki/etcd-overlay", k.etcdOverlayPKI.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
	}
	if exp, act := "test-cluster/pki/k8s", k.kubernetesPKI.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
	}
	if exp, act := "test-cluster/generic", k.secretsGeneric.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
	}
}
