// Copyright Jetstack Ltd. See LICENSE for details.
package api

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jetstack/vault-helper/pkg/kubernetes"
)

func TestDelete_Backend(t *testing.T) {
	Must(k.Ensure(), t)
	Must(k.Delete(), t)
	checkDryRun(true, t)

	for _, isNil := range []bool{true, false} {

		for _, b := range []kubernetes.Backend{
			kubernetes.NewPKI(k, "etcd-k8s", k.Log),
			kubernetes.NewPKI(k, "etcd-overlay", k.Log),
			kubernetes.NewPKI(k, "k8s", k.Log),
			kubernetes.NewPKI(k, "k8s-api-proxy", k.Log),
			k.NewGeneric(k.Log),
		} {

			mounts, err := v.Client().Sys().ListMounts()
			Must(err, t)

			mount, ok := mounts[vaultPath(b.Path())]
			if !isNil && (!ok || mount == nil) {
				t.Errorf("expcted to find mount '%s', did not", b.Path())
			}

			if isNil && ok && mount != nil {
				t.Errorf("expcted to not mount '%s', did", b.Path())
			}
		}

		Must(k.Ensure(), t)
	}

}

func TestDelete_EtcdRole(t *testing.T) {
	Must(k.Ensure(), t)
	Must(k.Delete(), t)
	checkDryRun(true, t)

	for _, isNil := range []bool{true, false} {
		for _, b := range []kubernetes.Backend{
			kubernetes.NewPKI(k, "etcd-k8s", k.Log),
			kubernetes.NewPKI(k, "etcd-overlay", k.Log),
		} {

			for _, role := range []string{"server", "client"} {

				path := filepath.Join(b.Path(), "roles", role)
				secret, err := v.Client().Logical().Read(path)
				Must(err, t)
				MustSecret(secret, isNil, t)
			}

		}

		Must(k.Ensure(), t)
	}
}

func TestDelete_KubernetesRole(t *testing.T) {
	Must(k.Ensure(), t)
	Must(k.Delete(), t)
	checkDryRun(true, t)

	b := kubernetes.NewPKI(k, "k8s", k.Log)
	for _, isNil := range []bool{true, false} {
		for _, role := range []string{
			"admin",
			"kube-apiserver",
			"kube-scheduler",
			"kube-controller-manager",
			"kube-proxy",
			"kubelet",
		} {

			path := filepath.Join(b.Path(), "roles", role)
			secret, err := v.Client().Logical().Read(path)
			Must(err, t)
			MustSecret(secret, isNil, t)
		}
		Must(k.Ensure(), t)
	}
}

func TestDelte_KubernetesAPIRole(t *testing.T) {
	Must(k.Ensure(), t)
	Must(k.Delete(), t)
	checkDryRun(true, t)

	b := kubernetes.NewPKI(k, "k8s-api-proxy", k.Log)
	for _, isNil := range []bool{true, false} {

		path := filepath.Join(b.Path(), "roles", "kube-apiserver")
		secret, err := v.Client().Logical().Read(path)
		Must(err, t)
		MustSecret(secret, isNil, t)

		Must(k.Ensure(), t)
	}
}

func TestDelete_Policies(t *testing.T) {
	Must(k.Ensure(), t)
	Must(k.Delete(), t)
	checkDryRun(true, t)

	for _, isNil := range []bool{true, false} {
		for _, role := range []string{"etcd", "master", "worker"} {
			policyName := fmt.Sprintf("%s/%s", clusterName, role)
			policy, err := v.Client().Sys().GetPolicy(policyName)
			Must(err, t)
			MustPolicy(policy, isNil, t)
		}

		Must(k.Ensure(), t)
	}

}

func TestDelete_InitToken(t *testing.T) {
	Must(k.Ensure(), t)
	Must(k.Delete(), t)
	checkDryRun(true, t)

	for _, isNil := range []bool{true, false} {
		for _, token := range k.NewInitTokens() {

			secret, err := v.Client().Logical().Read(token.Path())
			Must(err, t)
			MustSecret(secret, isNil, t)

			policyName := fmt.Sprintf("%s/%s-creator", clusterName, token.Role)
			policy, err := v.Client().Sys().GetPolicy(policyName)
			Must(err, t)
			MustPolicy(policy, isNil, t)
		}

		Must(k.Ensure(), t)
	}
}
