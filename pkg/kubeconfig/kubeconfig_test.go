package kubeconfig

import (
	"io/ioutil"
	"os"
	"os/user"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/jetstack-experimental/vault-helper/pkg/cert"
	"github.com/jetstack-experimental/vault-helper/pkg/instanceToken"
	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"
	"github.com/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

var vaultDev *vault_dev.VaultDev

var tempDirs []string

func TestMain(m *testing.M) {
	vaultDev = initVaultDev()

	// this runs all tests
	returnCode := m.Run()

	// shutdown vault
	vaultDev.Stop()

	// clean up tempdirs
	for _, dir := range tempDirs {
		os.RemoveAll(dir)
	}

	// return exit code according to the test runs
	os.Exit(returnCode)
}

// Test permissons of created files
func TestKubeconf_File_Perms(t *testing.T) {

	k := initKubernetes(t, vaultDev)
	c, i := initCert(t, vaultDev)

	token := k.InitTokens()["master"]
	if err := i.WriteTokenFile(i.InitTokenFilePath(), token); err != nil {
		t.Fatalf("error setting token for test: %s", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("error runinning cert: %s", err)
	}

	u := initKubeconf(t, vaultDev, c)

	if err := u.RunKube(); err != nil {
		t.Fatalf("error runinning kubeconfig: %s", err)
	}
}

// Check permissions of a file
func checkFilePerm(t *testing.T, path string, mode os.FileMode) {
	if fi, err := os.Stat(path); err != nil {
		t.Fatalf("error finding stats of '%s': %s", path, err)
	} else if fi.IsDir() {
		t.Fatalf("file should not be directory %s", path)
	} else if perm := fi.Mode(); perm != mode {
		t.Fatalf("destination has incorrect file permissons. exp=%s got=%s", mode, perm)
	}

}

// Init kubernetes for testing
func initKubernetes(t *testing.T, vaultDev *vault_dev.VaultDev) *kubernetes.Kubernetes {
	k := kubernetes.New(vaultDev.Client())
	k.SetClusterID("test-cluster")

	if err := k.Ensure(); err != nil {
		t.Fatalf("failed to ensure kubernetes: %s", err)
	}

	return k
}

// Start vault_dev for testing
func initVaultDev() *vault_dev.VaultDev {
	vaultDev := vault_dev.New()

	if err := vaultDev.Start(); err != nil {
		logrus.Fatalf("unable to initialise vault dev server for integration tests: %s", err)
	}

	return vaultDev
}

// Init Cert for tesing
func initCert(t *testing.T, vaultDev *vault_dev.VaultDev) (c *cert.Cert, i *instanceToken.InstanceToken) {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	log := logrus.NewEntry(logger)

	c = cert.New(vaultDev.Client(), log)
	c.SetRole("test-cluster/pki/k8s/sign/kube-apiserver")
	c.SetCommonName("k8s")
	c.SetBitSize(2048)

	if usr, err := user.Current(); err != nil {
		t.Fatalf("error getting info on current user: %s", err)
	} else {
		c.SetOwner(usr.Username)
		c.SetGroup(usr.Username)
	}

	// setup temporary directory for tests
	dir, err := ioutil.TempDir("", "test-cluster-dir")
	if err != nil {
		t.Fatal(err)
	}
	tempDirs = append(tempDirs, dir)
	c.SetVaultConfigPath(dir)
	c.SetDestination(dir + "/test")

	i = initInstanceToken(t, vaultDev, dir)

	return c, i
}

// Init Kubeconfig for tesing
func initKubeconf(t *testing.T, vaultDev *vault_dev.VaultDev, cert *cert.Cert) (u *Kubeconfig) {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	log := logrus.NewEntry(logger)

	u = New(vaultDev.Client(), log)
	u.SetCert(cert)
	u.SetFilePath(cert.Destination())

	return u
}

// Init instance token for testing
func initInstanceToken(t *testing.T, vaultDev *vault_dev.VaultDev, dir string) *instanceToken.InstanceToken {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	log := logrus.NewEntry(logger)

	i := instanceToken.New(vaultDev.Client(), log)
	i.SetRole("")
	i.SetClusterID("test-cluster")

	i.SetVaultConfigPath(dir)

	if _, err := os.Stat(i.InitTokenFilePath()); os.IsNotExist(err) {
		ifile, err := os.Create(i.InitTokenFilePath())
		if err != nil {
			t.Fatalf("%s", err)
		}
		defer ifile.Close()
	}

	_, err := os.Stat(i.TokenFilePath())
	if os.IsNotExist(err) {
		tfile, err := os.Create(i.TokenFilePath())
		if err != nil {
			t.Fatalf("%s", err)
		}
		defer tfile.Close()
	}

	i.WipeTokenFile(i.InitTokenFilePath())
	i.WipeTokenFile(i.TokenFilePath())

	return i
}
