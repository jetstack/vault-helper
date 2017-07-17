package kubernetes

import (
	"testing"
	"time"

	//"github.com/Sirupsen/logrus"
	//"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/kubernetes"

	vault_testing "github.com/hashicorp/vault/api"

	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

func TestPKI_Ensure(t *testing.T) {
	vault := vault_dev.New()
	if err := vault.Start(); err != nil {
		t.Skip("unable to initialise vault dev server for integration tests: ", err)
	}
	defer vault.Stop()

	k, err := New(RealVaultFromAPI(vault.Client()), "test-cluster-inside")
	if err != nil {
		t.Error("unexpected error", err)
		return
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

	k.etcdKubernetesPKI.DefaultLeaseTTL = time.Hour * 0
	k.etcdOverlayPKI.MaxLeaseTTL = time.Hour * 0
	k.kubernetesPKI.DefaultLeaseTTL = time.Hour * 0
	k.Ensure()
	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	basePath := k.clusterID + "/pki"
	policy_name := k.clusterID + "/" + "master"

	exists, err := k.etcdKubernetesPKI.getTokenPolicyExists(policy_name)
	if err != nil {
		t.Error("Error finding policy: ", err)
	}
	if exists {
		t.Error("Policy Found - it should not be")
	}

	rule := "\npath \"" + basePath + "/" + "etcd-overlay/sign/client" + "\" {\n    capabilities = [\"create\",\"read\",\"update\"]\n}\n"
	policy := k.NewPolicy(policy_name, rule, "master")

	err = policy.WritePolicy()
	if err != nil {
		t.Error("Error writting policy: ", err)
	}

	exists, err = k.etcdKubernetesPKI.getTokenPolicyExists(policy_name)
	if err != nil {
		t.Error("Error finding policy: ", err)
	}
	if !exists {
		t.Error("Policy not found")
	}

	pkiWrongType := NewPKI(k, "wrong-type-pki")

	err = k.vaultClient.Sys().Mount(
		k.Path()+"/pki/"+"wrong-type-pki",
		&vault_testing.MountInput{
			Description: "Kubernetes " + k.clusterID + "/" + "wrong-type-pki" + " CA",
			Type:        "generic",
			Config:      k.etcdKubernetesPKI.getMountConfigInput(),
		},
	)
	if err != nil {
		t.Error("Error Mounting: ", err)
	}

	err = pkiWrongType.Ensure()
	if err == nil {
		t.Error("Should have error from wrong type")
	}

}
