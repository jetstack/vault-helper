package kubernetes

import (
	//"testing"

	"github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"
	//"github.com/jetstack-experimental/vault-helper/pkg/mocks"
	//"github.com/Sirupsen/logrus"
)

type fakeVault struct {
	fakeVault   *MockVault
	fakeSys     *MockVaultSys
	fakeLogical *MockVaultLogical
	fakeAuth    *MockVaultAuth
}

func NewFakeVault(ctrl *gomock.Controller) *fakeVault {
	v := &fakeVault{
		fakeVault:   NewMockVault(ctrl),
		fakeSys:     NewMockVaultSys(ctrl),
		fakeLogical: NewMockVaultLogical(ctrl),
	}

	v.fakeVault.EXPECT().Sys().AnyTimes().Return(v.fakeSys)
	v.fakeVault.EXPECT().Logical().AnyTimes().Return(v.fakeLogical)
	v.fakeVault.EXPECT().Auth().AnyTimes().Return(v.fakeAuth)

	return v

}

func DoubleEnsure_fake(v *fakeVault) {

	mountInput1 := &vault.MountInput{
		Description: "Kubernetes test-cluster-inside/etcd-k8s CA",
		Type:        "pki",
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: "175320",
			MaxLeaseTTL:     "175320",
		},
	}

	mountInput2 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + "/" + "etcd-overlay" + " CA",
		Type:        "pki",
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: "175320",
			MaxLeaseTTL:     "175320",
		},
	}

	mountInput3 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + "/" + "k8s" + " CA",
		Type:        "pki",
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: "175320",
			MaxLeaseTTL:     "175320",
		},
	}

	v.fakeSys.EXPECT().ListMounts().AnyTimes().Return(nil, nil)

	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-k8s", mountInput1).Times(2).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-overlay", mountInput2).Times(2).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/k8s", mountInput3).Times(2).Return(nil)

}

func NewPolicy_fake(v *fakeVault) {
	policyName := "test-cluster-inside/master"
	policyRules := "path \"test-cluster-inside/pki/k8s/sign/kube-apiserver\" {\n        capabilities = [\"create\",\"read\",\"update\"]\n    }\n    "
	role := "master"
	clusterID := "test-cluster-inside"
	v.fakeSys.EXPECT().PutPolicy(policyName, policyRules).Times(1).Return(nil)

	createrRule := "path \"auth/token/create/" + clusterID + "-" + role + "+\" {\n    capabilities = [\"create\",\"read\",\"update\"]\n}"
	v.fakeSys.EXPECT().PutPolicy(policyName+"-creator", createrRule).Times(1).Return(nil)
}

func NewToken_fake(v *fakeVault) {

	rolePath := "auth/token/roles/test-cluster-inside-admin"
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
	v.fakeLogical.EXPECT().Write(rolePath, writeData).Times(1).Return(nil, nil)

	rolePath = "auth/token/roles/test-cluster-inside-kube-scheduler"
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
	v.fakeLogical.EXPECT().Write(rolePath, writeData).Times(1).Return(nil, nil)

}
