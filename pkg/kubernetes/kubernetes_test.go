// Copyright Jetstack Ltd. See LICENSE for details.
package kubernetes

import (
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/jetstack/vault-helper/pkg/testing/vault_dev"
)

const (
	clusterName = "test-cluster"
)

var (
	vaultDev *vault_dev.VaultDev
	k        *Kubernetes
)

func TestMain(m *testing.M) {
	v, err := vault_dev.InitVaultDev("../../bin/vault")
	if err != nil {
		logrus.Fatalf("failed to initiate vault for testing: %v", err)
	}
	vaultDev = v
	defer v.Stop()
	logrus.RegisterExitHandler(v.Stop)

	k8s := New(v.Client(), logrus.NewEntry(logrus.New()))
	k8s.SetClusterID(clusterName)
	k = k8s

	if k.Ensure(); err != nil {
		logrus.Fatalf("error ensuring: %v", err)
	}
	exitCode := m.Run()

	v.Stop()
	os.Exit(exitCode)
}

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
		t.Errorf("error '%v' should contain '%s'", err, msg)
	}

	err = isValidClusterID("invalid.cluster")
	if err == nil {
		t.Error("expected an error")
	} else if msg := "Invalid cluster ID"; !strings.Contains(err.Error(), msg) {
		t.Errorf("error '%v' should contain '%s'", err, msg)
	}
}

func TestKubernetes_Ensure(t *testing.T) {
	err := k.Ensure()
	if err != nil {
		t.Errorf("error ensuring kubernetes: %v", err)
	}
}

func TestKubernetes_NewToken_Role(t *testing.T) {
	if err := k.kubernetesBackend.WriteRole(k.k8sAdminRole()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	kubeSchedulerRole := k.k8sComponentRole("kube-scheduler")
	if err := k.kubernetesBackend.WriteRole(kubeSchedulerRole); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
