package kubernetes

import (
	"strings"
	"testing"
)

func TestIsValidClusterID(t *testing.T) {
	var err error

	err = isValidClusterID("valid-cluster")
	if err != nil {
		t.Error("unexpected an error: ", err)
	}

	err = isValidClusterID("valid-cluster01")
	if err != nil {
		t.Error("unexpected an error: ", err)
	}

	err = isValidClusterID("")
	if err == nil {
		t.Error("expected an error")
	} else if msg := "Invalid cluster ID"; !strings.Contains(err.Error(), msg) {
		t.Errorf("error '%s' should contain '%s'", err, msg)
	}

	err = isValidClusterID("invalid.cluster")
	if err == nil {
		t.Error("expected an error")
	} else if msg := "Invalid cluster ID"; !strings.Contains(err.Error(), msg) {
		t.Errorf("error '%s' should contain '%s'", err, msg)
	}

}

//go test -coverprofile=coverage.out
//  go tool cover -html=coverage.out

//func TestKubernetes_Double_Ensure(t *testing.T) {
//	vault := NewFakeVault(t)
//	defer vault.Finish()
//	k := vault.Kubernetes()
//
//	//vault.DoubleEnsure()
//
//	err := k.Ensure()
//	if err != nil {
//		t.Error("error ensuring: ", err)
//		return
//	}
//
//	err = k.Ensure()
//	if err != nil {
//		t.Error("error double ensuring: ", err)
//		return
//	}
//
//}

//func TestKubernetes_NewToken_Role(t *testing.T) {
//	vault := NewFakeVault(t)
//	defer vault.Finish()
//	k := vault.Kubernetes()
//
//	adminRole := k.k8sAdminRole()
//
//	err := k.kubernetesPKI.WriteRole(adminRole)
//
//	if err != nil {
//		t.Error("unexpected error", err)
//		return
//	}
//
//	kubeSchedulerRole := k.k8sComponentRole("kube-scheduler")
//
//	err = k.kubernetesPKI.WriteRole(kubeSchedulerRole)
//
//	if err != nil {
//		t.Error("unexpected error", err)
//		return
//	}
//
//}
