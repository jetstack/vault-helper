package cert

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/Sirupsen/logrus"
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
func TestCert_File_Perms(t *testing.T) {

	k := initKubernetes(t, vaultDev)
	c, i := initCert(t, vaultDev)

	token := k.InitTokens()["master"]
	if err := i.WriteTokenFile(i.InitTokenFilePath(), token); err != nil {
		t.Fatalf("Error setting token for test: \n%s", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("Error runinning cert:\n%s", err)
	}

	dir := filepath.Dir(c.Destination())
	if fi, err := os.Stat(dir); err != nil {
		t.Fatalf("Error finding stats of '%s':\n%s", dir, err)
	} else if !fi.IsDir() {
		t.Fatalf("Destination should be directory %s. It is not", dir)
	} else if perm := fi.Mode(); perm.String() != "drwxr-xr-x" {
		t.Fatalf("Destination has incorrect file permissons. Exp=drwxr-xr-x Got=%s", perm)
	}

	keyPem := filepath.Clean(c.Destination() + "-key.pem")
	dotPem := filepath.Clean(c.Destination() + ".pem")
	caPem := filepath.Clean(c.Destination() + "-ca.pem")
	checkFilePerm(t, keyPem, os.FileMode(0600))
	checkFilePerm(t, dotPem, os.FileMode(0644))
	checkFilePerm(t, caPem, os.FileMode(0644))
}

// Check permissions of a file
func checkFilePerm(t *testing.T, path string, mode os.FileMode) {
	if fi, err := os.Stat(path); err != nil {
		t.Fatalf("Error finding stats of '%s':\n%s", path, err)
	} else if fi.IsDir() {
		t.Fatalf("File should not be directory %s. It is.", path)
	} else if perm := fi.Mode(); perm != mode {
		t.Fatalf("Destination has incorrect file permissons. Exp=%s Got=%s", mode, perm)
	}

}

// Verify CAs exist
func TestCert_Verify_CA(t *testing.T) {

	k := initKubernetes(t, vaultDev)
	c, i := initCert(t, vaultDev)
	token := k.InitTokens()["master"]
	if err := i.WriteTokenFile(i.InitTokenFilePath(), token); err != nil {
		t.Fatalf("Error setting token for test: \n%s", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("Error runinning cert:\n%s", err)
	}

	dotPem := filepath.Clean(c.Destination() + ".pem")
	dat, err := ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", dotPem, err)
	}
	if dat == nil {
		t.Fatalf("No certificate at file '%s'. Expected certificate", dotPem)
	}

	caPem := filepath.Clean(c.Destination() + "-ca.pem")
	dat, err = ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", caPem, err)
	}
	if dat == nil {
		t.Fatalf("No certificate at file '%s'. Expected certificate", dotPem)
	}
}

// Test config file path
func TestCert_ConfigPath(t *testing.T) {
	k := initKubernetes(t, vaultDev)

	dir, err := ioutil.TempDir("", "test-cluster-dir")
	if err != nil {
		t.Fatal(err)
	}

	c, i := initCert(t, vaultDev)
	i.SetVaultConfigPath(dir)
	c.SetVaultConfigPath(dir)
	token := k.InitTokens()["master"]
	if err := i.WriteTokenFile(i.InitTokenFilePath(), token); err != nil {
		t.Fatalf("Error setting token for test: \n%s", err)
	}

	dotPem := filepath.Clean(c.Destination() + ".pem")
	if _, err := os.Stat(dotPem); !os.IsNotExist(err) {
		t.Fatalf("Expexted error 'File doesn't exist on file '.pem''. Instead:\n%s", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("Error runinning cert:\n%s", err)
	}

	caPem := filepath.Clean(c.Destination() + "-ca.pem")
	if _, err := os.Stat(caPem); err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", caPem, err)
	}

	dat, err := ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", dotPem, err)
	}
	if dat == nil {
		t.Fatalf("No certificate at file '%s'. Expected certificate", dotPem)
	}

	caPem = filepath.Clean(c.Destination() + "-ca.pem")
	dat, err = ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", caPem, err)
	}
	if dat == nil {
		t.Fatalf("No certificate at file '%s'. Expected certificate", dotPem)
	}
}

// Test if already existing valid certificate and key, they are kept
func TestCert_Exist_NoChange(t *testing.T) {
	k := initKubernetes(t, vaultDev)

	dir, err := ioutil.TempDir("", "test-cluster-dir")
	if err != nil {
		t.Fatal(err)
	}

	c, i := initCert(t, vaultDev)
	i.SetVaultConfigPath(dir)
	c.SetVaultConfigPath(dir)
	token := k.InitTokens()["master"]
	if err := i.WriteTokenFile(i.InitTokenFilePath(), token); err != nil {
		t.Fatalf("Error setting token for test: \n%s", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("Error running  cert:\n%s", err)
	}

	dotPem := filepath.Clean(c.Destination() + ".pem")
	datDotPem, err := ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", dotPem, err)
	}
	if datDotPem == nil {
		t.Fatalf("No certificate at file '%s'. Expected certificate", dotPem)
	}

	caPem := filepath.Clean(c.Destination() + "-ca.pem")
	datCAPem, err := ioutil.ReadFile(caPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", caPem, err)
	}
	if datCAPem == nil {
		t.Fatalf("No certificate at file '%s'. Expected certificate", dotPem)
	}

	keyPem := filepath.Clean(c.Destination() + "-key.pem")
	datKeyPem, err := ioutil.ReadFile(keyPem)
	if err != nil {
		t.Fatalf("Error reading from key file path: '%s':\n%s", keyPem, err)
	}
	if datKeyPem == nil {
		t.Fatalf("No key at file '%s'. Expected key", keyPem)
	}

	c.Log.Infof("-- Second run call --")
	if err := c.RunCert(); err != nil {
		t.Fatalf("Error running  cert:\n%s", err)
	}

	datDotPemAfter, err := ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", dotPem, err)
	}

	if string(datDotPem) != string(datDotPemAfter) {
		t.Fatalf("Certificate has been changed after cert call even though it exists. It shouldn't. %s", dotPem)
	}

	datCAPemAfter, err := ioutil.ReadFile(caPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", caPem, err)
	}
	if string(datCAPem) != string(datCAPemAfter) {
		t.Fatalf("Certificate has been changed after cert call even though it exists. It shouldn't. %s", caPem)
	}

	datKeyPemAfter, err := ioutil.ReadFile(keyPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", keyPem, err)
	}
	if string(datKeyPem) != string(datKeyPemAfter) {
		t.Fatalf("Key has been changed after cert call even though it exists. It shouldn't. %s", keyPem)
	}
}

func TestCert_Busy_Vault(t *testing.T) {
	k := initKubernetes(t, vaultDev)

	dir, err := ioutil.TempDir("", "test-cluster-dir")
	if err != nil {
		t.Fatal(err)
	}

	c, i := initCert(t, vaultDev)
	i.SetVaultConfigPath(dir)
	c.SetVaultConfigPath(dir)
	token := k.InitTokens()["master"]
	if err := i.WriteTokenFile(i.InitTokenFilePath(), token); err != nil {
		t.Fatalf("Error setting token for test: \n%s", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("Error running  cert:\n%s", err)
	}

	dotPem := filepath.Clean(c.Destination() + ".pem")
	datDotPem, err := ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", dotPem, err)
	}
	if datDotPem == nil {
		t.Fatalf("No certificate at file '%s'. Expected certificate", dotPem)
	}

	caPem := filepath.Clean(c.Destination() + "-ca.pem")
	datCAPem, err := ioutil.ReadFile(caPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", caPem, err)
	}
	if datCAPem == nil {
		t.Fatalf("No certificate at file '%s'. Expected certificate", dotPem)
	}

	keyPem := filepath.Clean(c.Destination() + "-key.pem")
	datKeyPem, err := ioutil.ReadFile(keyPem)
	if err != nil {
		t.Fatalf("Error reading from key file path: '%s':\n%s", keyPem, err)
	}
	if datKeyPem == nil {
		t.Fatalf("No key at file '%s'. Expected key", keyPem)
	}

	c.Log.Infof("-- Second run call --")
	c.vaultClient.SetToken("foo-bar")
	if err := c.RunCert(); err == nil {
		t.Fatalf("Expected 400 error, premisson denied")
	}

	datDotPemAfter, err := ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", dotPem, err)
	}

	if string(datDotPem) != string(datDotPemAfter) {
		t.Fatalf("Certificate has been changed after cert call even though it exists. It shouldn't. %s", dotPem)
	}

	datCAPemAfter, err := ioutil.ReadFile(caPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", caPem, err)
	}
	if string(datCAPem) != string(datCAPemAfter) {
		t.Fatalf("Certificate has been changed after cert call even though it exists. It shouldn't. %s", caPem)
	}

	datKeyPemAfter, err := ioutil.ReadFile(keyPem)
	if err != nil {
		t.Fatalf("Error reading from certificate file path: '%s':\n%s", keyPem, err)
	}
	if string(datKeyPem) != string(datKeyPemAfter) {
		t.Fatalf("Key has been changed after cert call even though it exists. It shouldn't. %s", keyPem)
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
	c.SetBitSize(2048)

	if usr, err := user.Current(); err != nil {
		t.Fatalf("Error getting info on current user: \n%s", err)
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
