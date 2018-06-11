// Copyright Jetstack Ltd. See LICENSE for details.
package api

import (
	"fmt"
	"path/filepath"
	"testing"

	vault "github.com/hashicorp/vault/api"

	"github.com/jetstack/vault-helper/pkg/kubernetes"
)

func TestDryRun_Backend(t *testing.T) {
	checkDryRun(false, t)

	for _, b := range []kubernetes.Backend{
		kubernetes.NewPKI(k, "etcd-k8s", k.Log),
		kubernetes.NewPKI(k, "etcd-overlay", k.Log),
		kubernetes.NewPKI(k, "k8s", k.Log),
		kubernetes.NewPKI(k, "k8s-api-proxy", k.Log),
		k.NewGeneric(k.Log),
	} {
		mount := &vault.MountInput{
			Type: b.Type(),
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

	for _, b := range []kubernetes.Backend{
		kubernetes.NewPKI(k, "etcd-k8s", k.Log),
		kubernetes.NewPKI(k, "etcd-overlay", k.Log),
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

	b := kubernetes.NewPKI(k, "k8s", k.Log)

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

	b := kubernetes.NewPKI(k, "k8s-api-proxy", k.Log)

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
}

func TestDryRun_Policies(t *testing.T) {
	Must(k.Ensure(), t)
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
	Must(k.Ensure(), t)
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
