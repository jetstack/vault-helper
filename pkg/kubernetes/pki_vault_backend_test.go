// Copyright Jetstack Ltd. See LICENSE for details.
package kubernetes

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

func TestPKIVaultBackend_Ensure(t *testing.T) {
	fv := NewFakeVault(t)
	defer fv.Finish()
	fv.ExpectWrite()

	fk := fv.Kubernetes()
	fv.PKIEnsure()

	if exp, act := fmt.Sprintf("%s-inside/pki/etcd-k8s", clusterName), fk.etcdKubernetesBackend.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
	}
	if exp, act := fmt.Sprintf("%s-inside/pki/etcd-overlay", clusterName), fk.etcdOverlayBackend.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
	}
	if exp, act := fmt.Sprintf("%s-inside/pki/k8s", clusterName), fk.kubernetesBackend.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
	}
	if exp, act := fmt.Sprintf("%s-inside/secrets", clusterName), fk.secretsGeneric.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
	}

	fk.etcdKubernetesBackend.DefaultLeaseTTL = time.Hour * 0
	fk.etcdOverlayBackend.MaxLeaseTTL = time.Hour * 0
	fk.kubernetesBackend.DefaultLeaseTTL = time.Hour * 0
	if err := fk.Ensure(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	policy_name := filepath.Join(fk.clusterID, "master")
	exists, err := fk.etcdKubernetesBackend.getTokenPolicyExists(policy_name)
	if err != nil {
		t.Errorf("failed to find policy: %v", err)
	}
	if exists {
		t.Errorf("unexpected policy found: %s", policy_name)
	}

	if err := fk.WritePolicy(fk.masterPolicy()); err != nil {
		t.Errorf("failed to write policy: %v", err)
	}

	exists, err = fk.etcdKubernetesBackend.getTokenPolicyExists(policy_name)
	if err != nil {
		t.Errorf("faileds to find policy: %v", err)
	}
	if !exists {
		t.Error("policy not found")
	}

	pkiWrongType := NewPKIVaultBackend(fk, "wrong-type-pki", logrus.NewEntry(logrus.New()))

	if err := fk.vaultClient.Sys().Mount(
		fk.Path()+"/pki/"+"wrong-type-pki",
		&vault.MountInput{
			Description: "Kubernetes " + fk.clusterID + "/" + "wrong-type-pki" + " CA",
			Type:        "generic",
			Config:      fk.etcdKubernetesBackend.getMountConfigInput(),
		},
	); err != nil {
		t.Errorf("failed to mount: %v", err)
	}

	if err = pkiWrongType.Ensure(); err == nil {
		t.Errorf("expected an error from wrong pki type")
	}
}
