package kubernetes

import (
	"testing"
)

func TestKubernetes_Backend_Path(t *testing.T) {
	k, err := New("test-cluster")
	if err != nil {
		t.Error("unexpected error", err)
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

	writeData := map[string]interface{}{
		"use_csr_common_name": false,
		"enforce_hostnames":   false,
		"organization":        "system:masters",
		"allowed_domains":     "admin",
		"allow_bare_domains":  true,
		"allow_localhost":     false,
		"allow_subdomains":    false,
		"allow_ip_sans":       false,
		"server_flag":         false,
		"client_flag":         true,
		"max_ttl":             "140h",
		"ttl":                 "140h",
	}

	err = WriteRoles(k.kubernetesPKI, writeData, "admin")

	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	writeData = map[string]interface{}{
		"use_csr_common_name": false,
		"enforce_hostnames":   false,
		"allowed_domains":     "kube-scheduler,system:kube-scheduler",
		"allow_bare_domains":  true,
		"allow_localhost":     false,
		"allow_subdomains":    false,
		"allow_ip_sans":       false,
		"server_flag":         false,
		"client_flag":         true,
		"max_ttl":             "140h",
		"ttl":                 "140h",
	}

	err = WriteRoles(k.kubernetesPKI, writeData, "kube-scheduler")

	if err != nil {
		t.Error("unexpected error", err)
		return
	}

}
