// Copyright Jetstack Ltd. See LICENSE for details.
package api

import (
	"fmt"
	"os"
	"testing"

	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/vault-helper/pkg/kubernetes"
	"github.com/jetstack/vault-helper/pkg/testing/vault_dev"
)

const (
	clusterName = "test-cluster"
)

var (
	v *vault_dev.VaultDev
	k *kubernetes.Kubernetes
)

func TestMain(m *testing.M) {
	vaultDev, err := vault_dev.InitVaultDev()
	if err != nil {
		logrus.Fatalf("failed to initiate vault for testing: %v", err)
	}
	v = vaultDev
	defer v.Stop()
	logrus.RegisterExitHandler(v.Stop)

	k8s := kubernetes.New(v.Client(), logrus.NewEntry(logrus.New()))
	k8s.SetClusterID(clusterName)
	k = k8s

	b, err := k.EnsureDryRun()
	if err != nil {
		logrus.Fatalf("unexpected error: %v", err)
	}
	if !b {
		logrus.Fatalf("expected changes required for dry run, got=none")
	}

	if k.Ensure(); err != nil {
		logrus.Fatalf("error ensuring: %v", err)
	}
	exitCode := m.Run()

	v.Stop()
	os.Exit(exitCode)
}

func Must(err error, t *testing.T) {
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func checkDryRun(exp bool, t *testing.T) {
	b, err := k.EnsureDryRun()
	Must(err, t)
	if b != exp {
		t.Errorf("unexpected changes required, exp=%t got=%t", exp, b)
	}
}

func createErrorData(dataMap map[string]interface{}) map[string]interface{} {
	for key, data := range map[string]interface{}{
		"max_ttl":         "0s",
		"ttl":             "0s",
		"organization":    "foo",
		"allowed_domains": []string{"foo"},
		"period":          "100s",
		"orphan":          "false",
	} {
		dataMap[key] = data
	}

	return dataMap
}

func MustSecret(secret *vault.Secret, isNil bool, t *testing.T) {
	if !isNil && secret == nil {
		t.Errorf("expected secret to not be nil, got=nil")
	} else if isNil && secret != nil {
		t.Errorf("expected secret to be nil, got=%+v", secret)
	}
}

func MustPolicy(policy string, isNil bool, t *testing.T) {
	if !isNil && policy == "" {
		t.Errorf("expected policy to not be nil, got=nil")
	} else if isNil && policy != "" {
		t.Errorf("expected policy to be nil, got=%s", policy)
	}
}

func vaultPath(path string) string {
	return fmt.Sprintf("%s/", path)
}
