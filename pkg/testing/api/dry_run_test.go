// Copyright Jetstack Ltd. See LICENSE for details.
package api

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	vault "github.com/hashicorp/vault/api"

	"github.com/jetstack/vault-helper/pkg/kubernetes"
)

func TestDryRun_BackendTypeDiffers(t *testing.T) {
	checkDryRun(false, t)

	for _, b := range []kubernetes.Backend{
		kubernetes.NewPKIVaultBackend(k, "etcd-k8s", k.Log),
		kubernetes.NewPKIVaultBackend(k, "etcd-overlay", k.Log),
		kubernetes.NewPKIVaultBackend(k, "k8s", k.Log),
		kubernetes.NewPKIVaultBackend(k, "k8s-api-proxy", k.Log),
		k.NewGenericVaultBackend(k.Log),
	} {

		//toggle the backend type
		backendType := "pki"
		if b.Type() == "pki" {
			backendType = "generic"
		}

		mount := &vault.MountInput{
			Type: backendType,
			Config: vault.MountConfigInput{
				DefaultLeaseTTL: "",
				MaxLeaseTTL:     "",
			},
		}

		Must(v.Client().Sys().Unmount(b.Path()), t)
		checkDryRun(true, t)

		Must(v.Client().Sys().Mount(b.Path(), mount), t)
		checkDryRun(true, t)

		Must(v.Client().Sys().Unmount(b.Path()), t)
		checkDryRun(true, t)

		Must(k.Ensure(), t)
		checkDryRun(false, t)
	}
}

func TestDryRun_EtcdRole(t *testing.T) {
	checkDryRun(false, t)

	for _, b := range []kubernetes.Backend{
		kubernetes.NewPKIVaultBackend(k, "etcd-k8s", k.Log),
		kubernetes.NewPKIVaultBackend(k, "etcd-overlay", k.Log),
	} {

		for _, role := range []string{"server", "client"} {
			Must(k.Ensure(), t)
			checkDryRun(false, t)

			path := filepath.Join(b.Path(), "roles", role)
			secret, err := v.Client().Logical().Read(path)
			Must(err, t)

			if secret == nil {
				t.Error("secret is nil, expected not nil")
			}

			_, err = v.Client().Logical().Write(path, createErrorData(secret.Data))
			Must(err, t)
			checkDryRun(true, t)
		}

	}

	Must(k.Ensure(), t)
	checkDryRun(false, t)
}

func TestDryRun_KubernetesRole(t *testing.T) {
	checkDryRun(false, t)

	b := kubernetes.NewPKIVaultBackend(k, "k8s", k.Log)

	for _, role := range []string{
		"admin",
		"kube-apiserver",
		"kube-scheduler",
		"kube-controller-manager",
		"kube-proxy",
		"kubelet",
	} {

		Must(k.Ensure(), t)
		checkDryRun(false, t)

		path := filepath.Join(b.Path(), "roles", role)
		secret, err := v.Client().Logical().Read(path)
		Must(err, t)

		if secret == nil {
			t.Error("secret is nil, expected not nil")
		}

		_, err = v.Client().Logical().Write(path, createErrorData(secret.Data))
		Must(err, t)
		checkDryRun(true, t)
	}

	Must(k.Ensure(), t)
	checkDryRun(false, t)
}

func TestDryRun_KubernetesAPIRole(t *testing.T) {
	checkDryRun(false, t)

	b := kubernetes.NewPKIVaultBackend(k, "k8s-api-proxy", k.Log)

	Must(k.Ensure(), t)
	checkDryRun(false, t)

	path := filepath.Join(b.Path(), "roles", "kube-apiserver")
	secret, err := v.Client().Logical().Read(path)
	Must(err, t)

	if secret == nil {
		t.Error("secret is nil, expected not nil")
	}

	_, err = v.Client().Logical().Write(path, createErrorData(secret.Data))
	Must(err, t)
	checkDryRun(true, t)

	Must(k.Ensure(), t)
	checkDryRun(false, t)
}

func TestDryRun_Policies(t *testing.T) {
	checkDryRun(false, t)

	policy := `path "test-cluster/pki/etcd-overlay/sign/server" {
	  capabilities = []
	}`

	for _, role := range []string{"etcd", "master", "worker"} {
		policyName := fmt.Sprintf("%s/%s", clusterName, role)
		Must(v.Client().Sys().PutPolicy(policyName, policy), t)
		checkDryRun(true, t)

		Must(k.Ensure(), t)
		checkDryRun(false, t)
	}
}

func TestDryRun_InitToken(t *testing.T) {
	checkDryRun(false, t)

	policy := `path "test-cluster/pki/etcd-overlay/sign/server" {
  capabilities = []
}`

	for _, token := range k.NewInitTokens() {

		secret, err := v.Client().Logical().Read(token.Path())
		Must(err, t)

		_, err = v.Client().Logical().Write(token.Path(), createErrorData(secret.Data))
		Must(err, t)
		checkDryRun(true, t)

		policyName := fmt.Sprintf("%s/%s-creator", clusterName, token.Role)
		Must(v.Client().Sys().PutPolicy(policyName, policy), t)
		checkDryRun(true, t)

		Must(k.Ensure(), t)
		checkDryRun(false, t)
	}
}

func TestDryRun_InitToken_Revoke(t *testing.T) {
	checkDryRun(false, t)

	for _, token := range k.NewInitTokens() {
		tokenS, err := token.InitToken()
		Must(err, t)

		_, err = v.Client().Auth().Token().Lookup(tokenS)
		Must(err, t)

		Must(v.Client().Auth().Token().RevokeOrphan(tokenS), t)
		checkDryRun(true, t)

		_, err = v.Client().Auth().Token().Lookup(tokenS)
		if err == nil {
			t.Fatal("expected error, got none")
		}

		if !strings.Contains(err.Error(), "Code: 403.") ||
			!strings.Contains(err.Error(), "bad token") {
			t.Fatalf("unexpected error: %s", err)
		}

		Must(k.Ensure(), t)
		checkDryRun(false, t)

		_, err = v.Client().Auth().Token().Lookup(tokenS)
		Must(err, t)
	}
}

func TestDryRun_MountTTL(t *testing.T) {
	checkDryRun(false, t)

	Must(v.Client().Sys().TuneMount("/auth/token", vault.MountConfigInput{
		MaxLeaseTTL: "5s",
	}), t)

	checkDryRun(true, t)

	Must(k.Ensure(), t)
	checkDryRun(false, t)
}
