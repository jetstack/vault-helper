package kubernetes

import (
	"testing"

	//"github.com/golang/mock/gomock"
	//vault "github.com/hashicorp/vault/api"
	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
	//"time"
)

//go test -coverprofile=coverage.out
//  go tool cover -html=coverage.out
func TestKubernetes_Run_Setup_Test(t *testing.T) {
	args := []string{"test-cluster-run"}
	Run(nil, args)
}

func TestInvalid_Cluster_ID(t *testing.T) {
	vault := vault_dev.New()

	if err := vault.Start(); err != nil {
		t.Skip("unable to initialise vault dev server for integration tests: ", err)
	}
	defer vault.Stop()

	_, err := New(RealVaultFromAPI(vault.Client()), "INVALID CLUSTER ID $^^%*$^")
	if err == nil {
		t.Error("Should be invalid vluster ID")
	}

	_, err = New(RealVaultFromAPI(vault.Client()), "5INVALID CLUSTER ID $^^%*$^")
	if err == nil {
		t.Error("Should be invalid vluster ID")
	}

}

func TestKubernetes_Double_Ensure(t *testing.T) {
	vault := vault_dev.New()

	if err := vault.Start(); err != nil {
		t.Skip("unable to initialise vault dev server for integration tests: ", err)
	}
	defer vault.Stop()

	k, err := New(RealVaultFromAPI(vault.Client()), "test-cluster-inside")
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

}
func TestKubernetes_NewPolicy_Role(t *testing.T) {
	vault := vault_dev.New()

	if err := vault.Start(); err != nil {
		t.Skip("unable to initialise vault dev server for integration tests: ", err)
	}
	defer vault.Stop()

	k, err := New(RealVaultFromAPI(vault.Client()), "test-cluster-inside")
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

	err = masterPolicy.CreateTokenCreater()
	if err != nil {
		t.Error("unexpected error", err)
	}

}

func TestKubernetes_NewToken_Role(t *testing.T) {

	vault := vault_dev.New()

	if err := vault.Start(); err != nil {
		t.Skip("unable to initialise vault dev server for integration tests: ", err)
	}
	defer vault.Stop()

	k, err := New(RealVaultFromAPI(vault.Client()), "test-cluster-inside")
	if err != nil {
		t.Error("unexpected error", err)
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

}
