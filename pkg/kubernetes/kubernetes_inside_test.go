package kubernetes

import (
	"testing"

	//"github.com/Sirupsen/logrus"
	"github.com/golang/mock/gomock"
	//vault "github.com/hashicorp/vault/api"
	//"github.com/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
	//"time"
)

//go test -coverprofile=coverage.out
//  go tool cover -html=coverage.out
func TestKubernetes_Run_Setup_Test(t *testing.T) {
	args := []string{"test-cluster-run"}
	Run(nil, args)
}

func TestInvalid_Cluster_ID(t *testing.T) {

	_, err := New(nil, "INVALID CLUSTER ID $^^%*$^")
	if err == nil {
		t.Error("Should be invalid vluster ID")
	}

	_, err = New(nil, "5INVALID CLUSTER ID $^^%*$^")
	if err == nil {
		t.Error("Should be invalid vluster ID")
	}

}

func TestKubernetes_Double_Ensure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	vault := NewFakeVault(mockCtrl)

	DoubleEnsure_fake(vault)

	k, err := New(nil, "test-cluster-inside")
	if err != nil {
		t.Error("unexpected error", err)
	}

	k.vaultClient = vault.fakeVault

	err = k.Ensure()
	if err != nil {
		t.Error("error ensuring: ", err)
		return
	}

	err = k.Ensure()
	if err != nil {
		t.Error("error double ensuring: ", err)
		return
	}

}

func TestKubernetes_NewPolicy_Role(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	vault := NewFakeVault(mockCtrl)

	NewPolicy_fake(vault)

	k, err := New(nil, "test-cluster-inside")
	if err != nil {
		t.Error("unexpected error", err)
	}

	k.vaultClient = vault.fakeVault

	policyName := "test-cluster-inside/master"
	policyRules := `
path "test-cluster-inside/pki/k8s/sign/kube-apiserver" {
	capabilities = ["create","read","update"]
}
`
	role := "master"

	masterPolicy := k.NewPolicy(policyName, policyRules, role)

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

}

func TestKubernetes_NewToken_Role(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	vault := NewFakeVault(mockCtrl)

	NewToken_fake(vault)

	k, err := New(nil, "test-cluster-inside")
	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	k.vaultClient = vault.fakeVault

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

	kubeSchedulerRole := k.NewTokenRole("kube-scheduler", writeData)

	err = kubeSchedulerRole.WriteTokenRole()

	if err != nil {
		t.Error("unexpected error", err)
		return
	}

}
