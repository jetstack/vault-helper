package kubernetes

import (
	"testing"

	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

//go test -coverprofile=coverage.out
//  go tool cover -html=coverage.out

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

	err = k.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
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

	adminRole := k.NewTokenRole("admin", writeData)
	err = adminRole.WriteTokenRole()

	if err != nil {
		t.Error("unexpected error", err)
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

	kubeSchedulerRole := k.NewTokenRole("kube-scheduler", writeData)

	err = kubeSchedulerRole.WriteTokenRole()

	if err != nil {
		t.Error("unexpected error", err)
	}

	policyName := "test-cluster-inside/master"
	policyRules := "path \"test-cluster-inside/pki/k8s/sign/kube-apiserver\" {\n        capabilities = [\"create\",\"read\",\"update\"]\n    }\n    "
	role := "master"

	masterPolicy := k.NewPolicy(policyName, policyRules, role)

	err = masterPolicy.WritePolicy()
	if err != nil {
		t.Error("unexpected error", err)
	}

	generic := k.NewGeneric()
	err = generic.Ensure()
	if err != nil {
		t.Error("unexpected error", err)
	}

	err = masterPolicy.CreateTokenCreater()
	if err != nil {
		t.Error("unexpected error", err)
	}

	masterToken := k.NewInitToken(policyName, role)
	err = masterToken.CreateToken()
	if err != nil {
		t.Error("unexpected error", err)
	}

	err = masterToken.WriteInitToken()
	if err != nil {
		t.Error("unexpected error", err)
	}

}
