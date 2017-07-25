package kubernetes

import (
	"testing"

	"github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"
	"strings"
)

type fakeVault struct {
	ctrl *gomock.Controller

	fakeVault   *MockVault
	fakeSys     *MockVaultSys
	fakeLogical *MockVaultLogical
	fakeAuth    *MockVaultAuth
	fakeToken   *MockVaultToken
}

func NewFakeVault(t *testing.T) *fakeVault {
	ctrl := gomock.NewController(t)

	v := &fakeVault{
		ctrl: ctrl,

		fakeVault:   NewMockVault(ctrl),
		fakeSys:     NewMockVaultSys(ctrl),
		fakeLogical: NewMockVaultLogical(ctrl),
		fakeAuth:    NewMockVaultAuth(ctrl),
		fakeToken:   NewMockVaultToken(ctrl),
	}

	v.fakeVault.EXPECT().Sys().AnyTimes().Return(v.fakeSys)
	v.fakeVault.EXPECT().Logical().AnyTimes().Return(v.fakeLogical)
	v.fakeVault.EXPECT().Auth().AnyTimes().Return(v.fakeAuth)
	v.fakeAuth.EXPECT().Token().AnyTimes().Return(v.fakeToken)

	return v
}

func (v *fakeVault) Kubernetes() *Kubernetes {
	k := New(nil)
	k.SetClusterID("test-cluster-inside")
	k.vaultClient = v.fakeVault
	return k
}

func (v *fakeVault) Finish() {
	v.ctrl.Finish()
}

func (v *fakeVault) DoubleEnsure() {

	mountInput1 := &vault.MountInput{
		Description: "Kubernetes test-cluster-inside/etcd-k8s CA",
		Type:        "pki",
	}

	mountInput2 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + "/" + "etcd-overlay" + " CA",
		Type:        "pki",
	}

	mountInput3 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + "/" + "k8s" + " CA",
		Type:        "pki",
	}

	mountInput4 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + " secrets",
		Type:        "generic",
	}

	description1 := "Kubernetes test-cluster-inside/etcd-k8s CA"
	data1 := map[string]interface{}{
		"common_name": description1,
		"ttl":         "630720000s",
	}

	description2 := "Kubernetes test-cluster-inside/etcd-overlay CA"
	data2 := map[string]interface{}{
		"common_name": description2,
		"ttl":         "630720000s",
	}

	description3 := "Kubernetes test-cluster-inside/k8s CA"
	data3 := map[string]interface{}{
		"common_name": description3,
		"ttl":         "630720000s",
	}

	data4 := map[string]interface{}{
		"allow_any_name":      true,
		"client_flag":         true,
		"use_csr_common_name": false,
		"use_csr_sans":        false,
		"max_ttl":             "2592000s",
		"ttl":                 "2592000s",
		"allow_ip_sans":       true,
		"server_flag":         false,
	}

	data5 := map[string]interface{}{
		"allow_any_name":      true,
		"client_flag":         true,
		"use_csr_common_name": false,
		"use_csr_sans":        false,
		"max_ttl":             "2592000s",
		"ttl":                 "2592000s",
		"allow_ip_sans":       true,
		"server_flag":         true,
	}

	data6 := map[string]interface{}{
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
		"max_ttl":             "31536000s",
		"ttl":                 "31536000s",
	}

	data7 := map[string]interface{}{
		"use_csr_common_name": false,
		"use_csr_sans":        false,
		"enforce_hostnames":   false,
		"allow_localhost":     true,
		"allow_any_name":      true,
		"allow_bare_domains":  true,
		"allow_ip_sans":       true,
		"server_flag":         true,
		"client_flag":         false,
		"max_ttl":             "2592000s",
		"ttl":                 "2592000s",
	}

	data8 := map[string]interface{}{
		"use_csr_common_name": false,
		"enforce_hostnames":   false,
		"allowed_domains":     "kube-scheduler,system:kube-scheduler",
		"allow_bare_domains":  true,
		"allow_localhost":     false,
		"allow_subdomains":    false,
		"allow_ip_sans":       false,
		"server_flag":         false,
		"client_flag":         true,
		"max_ttl":             "2592000s",
		"ttl":                 "2592000s",
	}

	data9 := map[string]interface{}{
		"use_csr_common_name": false,
		"enforce_hostnames":   false,
		"allowed_domains":     "kube-controller-manager,system:kube-controller-manager",
		"allow_bare_domains":  true,
		"allow_localhost":     false,
		"allow_subdomains":    false,
		"allow_ip_sans":       false,
		"server_flag":         false,
		"client_flag":         true,
		"max_ttl":             "2592000s",
		"ttl":                 "2592000s",
	}

	data10 := map[string]interface{}{
		"use_csr_common_name": false,
		"use_csr_sans":        false,
		"enforce_hostnames":   false,
		"organization":        "system:nodes",
		"allowed_domains":     strings.Join([]string{"kubelet", "system:node", "system:node:*"}, ","),
		"allow_bare_domains":  true,
		"allow_glob_domains":  true,
		"allow_localhost":     false,
		"allow_subdomains":    false,
		"server_flag":         true,
		"client_flag":         true,
		"max_ttl":             "2592000s",
		"ttl":                 "2592000s",
	}

	data11 := map[string]interface{}{
		"use_csr_common_name": false,
		"enforce_hostnames":   false,
		"allowed_domains":     "kube-proxy,system:kube-proxy",
		"allow_bare_domains":  true,
		"allow_localhost":     false,
		"allow_subdomains":    false,
		"allow_ip_sans":       false,
		"server_flag":         false,
		"client_flag":         true,
		"max_ttl":             "2592000s",
		"ttl":                 "2592000s",
	}

	v.fakeSys.EXPECT().ListMounts().AnyTimes().Return(nil, nil)

	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-k8s", mountInput1).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-overlay", mountInput2).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/k8s", mountInput3).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/secrets", mountInput4).Times(1).Return(nil)

	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/etcd-k8s/root/generate/internal", data1).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/etcd-overlay/root/generate/internal", data2).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/root/generate/internal", data3).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/secrets/service-accounts", gomock.Any()).Return(nil, nil)

	v.fakeLogical.EXPECT().Read("test-cluster-inside/pki/etcd-k8s/cert/ca").Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Read("test-cluster-inside/pki/etcd-overlay/cert/ca").Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Read("test-cluster-inside/pki/k8s/cert/ca").Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Read("test-cluster-inside/secrets/service-accounts").Times(1).Return(nil, nil)

	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/etcd-k8s/roles/client", data4).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/etcd-k8s/roles/server", data5).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/etcd-overlay/roles/client", data4).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/etcd-overlay/roles/server", data5).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/admin", data6).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/kube-apiserver", data7).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/kube-scheduler", data8).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/kube-controller-manager", data9).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/kubelet", data10).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/kube-proxy", data11).Times(1).Return(nil, nil)

	//v.fakeSys.EXPECT().PutPolicy("test-cluster-inside/etcd", gomock.Any()).Times(1).Return(nil)
	//v.fakeSys.EXPECT().PutPolicy("test-cluster-inside/master", gomock.Any()).Times(1).Return(nil)
	//v.fakeSys.EXPECT().PutPolicy("test-cluster-inside/worker", gomock.Any()).Times(1).Return(nil)

	//v.fakeLogical.EXPECT().Read("test-cluster-inside/secrets/init_token_etcd").AnyTimes().Return(nil, nil)

	//v.fakeLogical.EXPECT().Write("auth/token/roles/test-cluster-inside-etcd", gomock.Any()).AnyTimes().Return(nil, nil)

	//v.fakeAuth.EXPECT().Token().Times(1).Return(nil)
}

//func (v *fakeVault) NewPolicy() {
//	//	rules := `
//	//	path "test-cluster-inside/pki/etcd-k8s/sign/client" {
//	//	  capabilities = ["create", "read", "update"]
//	//	}
//	//
//	//	path "test-cluster-inside/secrets/service-accounts" {
//	//	  capabilities = ["read"]
//	//	}
//	//
//	//	path "test-cluster-inside/pki/k8s/sign/kube-apiserver" {
//	//	  capabilities = ["create", "read", "update"]
//	//	}
//	//
//	//	path "test-cluster-inside/pki/k8s/sign/kube-scheduler" {
//	//	  capabilities = ["create", "read", "update"]
//	//	}
//	//
//	//	path "test-cluster-inside/pki/k8s/sign/kube-controller-manager" {
//	//	  capabilities = ["create", "read", "update"]
//	//	}
//	//
//	//	path "test-cluster-inside/pki/k8s/sign/admin" {
//	//	  capabilities = ["create", "read", "update"]
//	//	}
//	//
//	//	path "test-cluster-inside/pki/k8s/sign/kubelet" {
//	//	  capabilities = ["create", "read", "update"]
//	//	}
//	//
//	//	path "test-cluster-inside/pki/k8s/sign/kube-proxy" {
//	//	  capabilities = ["create", "read", "update"]
//	//	}
//	//
//	//	path "test-cluster-inside/pki/etcd-overlay/sign/client" {
//	//	  capabilities = ["create", "read", "update"]
//	//	}]`
//	v.fakeSys.EXPECT().PutPolicy(gomock.Any, gomock.Any()).AnyTimes().Return(nil)
//}

func (v *fakeVault) NewToken() {

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

func (v *fakeVault) PKIEnsure() {

	mountInput1 := &vault.MountInput{
		Description: "Kubernetes test-cluster-inside/etcd-k8s CA",
		Type:        "pki",
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: "0",
			MaxLeaseTTL:     "630720000",
		},
	}

	mountInput2 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + "/" + "etcd-overlay" + " CA",
		Type:        "pki",
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: "630720000",
			MaxLeaseTTL:     "0",
		},
	}

	mountInput3 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + "/" + "k8s" + " CA",
		Type:        "pki",
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: "0",
			MaxLeaseTTL:     "630720000",
		},
	}
	v.fakeSys.EXPECT().ListMounts().AnyTimes().Return(nil, nil)

	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-k8s", mountInput1).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-overlay", mountInput2).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/k9s", mountInput3).Times(1).Return(nil)

	firstGet := v.fakeSys.EXPECT().GetPolicy("test-cluster-inside/master").Times(1).Return("", nil)
	v.fakeSys.EXPECT().GetPolicy("test-cluster-inside/master").Times(1).Return("true", nil).After(firstGet)

	policyName := "test-cluster-inside/master"
	policyRules := "\npath \"test-cluster-inside/pki/" + "etcd-overlay/sign/client" + "\" {\n    capabilities = [\"create\",\"read\",\"update\"]\n}\n"
	v.fakeSys.EXPECT().PutPolicy(policyName, policyRules).Times(1).Return(nil)
}
