package kubernetes

import (
	"testing"

	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

func TestKubernetes_Backend_Path(t *testing.T) {
	vault := vault_dev.New()
	if err := vault.Start(); err != nil {
		t.Skip("unable to initialise vault dev server for integration tests: ", err)
	}
	defer vault.Stop()

	k, err := New(vault.Client(), "test-cluster-inside")
	if err != nil {
		t.Error("unexpected error", err)
	}

	err = k.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
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

	adminRole := NewTokenRole("admin", writeData, k)
	err = adminRole.WriteTokenRole()

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

	kubeSchedulerRole := NewTokenRole("kube-scheduler", writeData, k)

	err = kubeSchedulerRole.WriteTokenRole()

	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	policyName := "test-cluster-inside/master"
	policyRules := "path \"test-cluster-inside/pki/k8s/sign/kube-apiserver\" {\n        capabilities = [\"create\",\"read\",\"update\"]\n    }\n    "
	role := "master"

	masterPolicy := NewPolicy(policyName, policyRules, role, k)

	err = masterPolicy.WritePolicy()
	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	err = masterPolicy.CreateTokenCreater()
	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	masterToken := NewInitToken(policyName, role, k)
	err = masterToken.CreateToken()
	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	//err = masterToken.WriteInitToken()
	//if err != nil {
	//	t.Error("unexpected error", err)
	//	return
	//}

}
