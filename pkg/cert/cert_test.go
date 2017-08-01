package cert

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/Sirupsen/logrus"
	//vault "github.com/hashicorp/vault/api"
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

func TestCert_Dirs_Perms(t *testing.T) {

	k := initKubernetes(t, vaultDev)
	c, i := initCert(t, vaultDev)

	token := k.InitTokens()["master"]
	if err := i.WriteTokenFile(i.InitTokenFilePath(), token); err != nil {
		t.Fatalf("Error setting token for test: \n%s", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("Error runinning cert:\n%s", err)
	}

	if fi, err := os.Stat(c.Destination()); err != nil {
		t.Fatalf("Error finding stats of '%s':\n%s", c.Destination(), err)
	} else if !fi.IsDir() {
		t.Fatalf("Destination should be directory %s. It is not", c.Destination())
	} else if perm := fi.Mode(); perm.String() != "drwxr-xr-x" {
		t.Fatalf("Destination has incorrect file permissons. Exp=drwxr-xr-x Got=%s", perm)
	}

	keyPem := filepath.Join(c.Destination(), "-key.pem")
	dotPem := filepath.Join(c.Destination(), ".pem")
	caPem := filepath.Join(c.Destination(), "-ca.pem")
	checkFilePerm(t, keyPem, os.FileMode(0600))
	checkFilePerm(t, dotPem, os.FileMode(0644))
	checkFilePerm(t, caPem, os.FileMode(0644))
}

func checkFilePerm(t *testing.T, path string, mode os.FileMode) {
	if fi, err := os.Stat(path); err != nil {
		t.Fatalf("Error finding stats of '%s':\n%s", path, err)
	} else if fi.IsDir() {
		t.Fatalf("File should not be directory %s. It is.", path)
	} else if perm := fi.Mode(); perm != mode {
		t.Fatalf("Destination has incorrect file permissons. Exp=%s Got=%s", mode, perm)
	}

}

// Init Cert for tesing
func initCert(t *testing.T, vaultDev *vault_dev.VaultDev) (c *Cert, i *instanceToken.InstanceToken) {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	log := logrus.NewEntry(logger)

	c = New(vaultDev.Client(), log)
	c.SetRole("test-cluster/pki/k8s/sign/kube-apiserver")
	c.SetCommonName("k8s")

	if usr, err := user.Current(); err != nil {
		t.Fatalf("Error getting current user: \n%s", err)
	} else {
		c.SetOwner(usr.Username)
		c.SetGroup("wheel") // Work out how to find this
	}
	//if grp, err := user.Current(); err != nil {
	//	t.Fatal("Error getting current user: \n%s", err)
	//} else {
	//	c.SetOwner(usr.Username)
	//}

	// setup temporary directory for tests
	dir, err := ioutil.TempDir("", "test-cluster-dir")
	if err != nil {
		t.Fatal(err)
	}
	tempDirs = append(tempDirs, dir)
	c.SetVaultConfigPath(dir)
	c.SetDestination(dir)

	i = initInstanceToken(t, vaultDev, dir)

	return c, i

}

// Init kubernetes for testing
func initKubernetes(t *testing.T, vaultDev *vault_dev.VaultDev) *kubernetes.Kubernetes {
	k := kubernetes.New(vaultDev.Client())
	k.SetClusterID("test-cluster")

	if err := k.Ensure(); err != nil {
		t.Fatalf("Error ensuring kubernetes: \n%s", err)
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
