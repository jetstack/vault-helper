package instanceToken_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
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

// Token exists at token_file - renew
func TestRenew_Token_Exists(t *testing.T) {

	k := initKubernetes(t, vaultDev)
	i := initInstanceToken(t, vaultDev)

	token := k.InitTokens()["master"]
	if err := i.WriteTokenFile(i.InitTokenFilePath(), token); err != nil {
		t.Fatalf("Error setting token for test: \n%s", err)
	}

	ttl, err := getTTL(vaultDev, token, i)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if err := i.TokenRenewRun(); err != nil {
		t.Fatalf("Error renewing token from token file (Exists): \n%s", err)
	}

	newttl, err := getTTL(vaultDev, token, i)
	if err != nil {
		t.Fatalf("%s", err)
	}

	i.Log.Debugf("old ttl: %ss    new ttl: %ss", strconv.Itoa(ttl), strconv.Itoa(newttl))

	if ttl > newttl {
		t.Fatalf("Token was not renewed - old ttl higher than new\nold=%s new=%s", strconv.Itoa(ttl), strconv.Itoa(newttl))
	}

	tokenCheckFiles(t, i)

	return
}

// Token doesn't exist at token file - generate a new form init_token file; renew token
func TestRenew_Token_NotExists(t *testing.T) {

	k := initKubernetes(t, vaultDev)
	i := initInstanceToken(t, vaultDev)

	if err := i.WriteTokenFile(i.InitTokenFilePath(), k.InitTokens()["master"]); err != nil {
		t.Fatalf("Error setting token for test: \n%s", err)
	}

	ttl, err := getTTL(vaultDev, i.Token(), i)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if err := i.TokenRenewRun(); err != nil {
		t.Fatalf("Error renewing token from token file (!Exist): \n%s", err)
	}

	newttl, err := getTTL(vaultDev, i.Token(), i)
	if err != nil {
		t.Fatalf("%s", err)
	}

	i.Log.Debugf("old ttl: %ss    new ttl: %ss", strconv.Itoa(ttl), strconv.Itoa(newttl))

	if ttl > newttl {
		t.Fatalf("Token was not renewed - old ttl higher than new\nold=%s new=%s", strconv.Itoa(ttl), strconv.Itoa(newttl))
	}

	tokenCheckFiles(t, i)

	return
}

// Token exists but can't be renewed - return error
func TestRenew_Token_Exists_NoRenew(t *testing.T) {

	initKubernetes(t, vaultDev)
	i := initInstanceToken(t, vaultDev)

	notRenewable := false
	tCreateRequest := &vault.TokenCreateRequest{
		DisplayName: "master",
		Policies:    []string{"root"},
		Renewable:   &notRenewable,
	}

	newToken, err := vaultDev.Client().Auth().Token().CreateOrphan(tCreateRequest)
	if err != nil {
		t.Fatalf("Unexpexted error creating unrenewable token:\n%s", err)
	}

	if err := i.WriteTokenFile(i.TokenFilePath(), newToken.Auth.ClientToken); err != nil {
		t.Fatalf("Error setting token for test: \n%s", err)
	}

	err = i.TokenRenewRun()
	i.Log.Debugf("%s", err)

	if err == nil {
		t.Fatalf("Expected an error - token not renewable. Fail")
	}

	if err.Error() == "Token not renewable: "+i.Token() {
		i.Log.Debugf("Error returned successfully - token is not renewable")
	} else {
		t.Errorf("Unexpected error. Fail.\n%s", err)
	}

	return
}

// Token doesn't exist at either file - return error
func TestRenew_Token_NeitherExist(t *testing.T) {

	initKubernetes(t, vaultDev)
	i := initInstanceToken(t, vaultDev)

	err := i.TokenRenewRun()

	if err == nil {
		t.Fatalf("Expected an error - init file is empty")
	}

	i.Log.Debugf("%s", err)
	str := "Error generating new token: \nInit token was not read from file: " + i.InitTokenFilePath()
	if err.Error() == str {
		i.Log.Debugf("Error returned successfully - no init token in file")
	} else {
		t.Errorf("Unexpected error. Fail.\n%s", err)
	}

	return
}

// Get ttl form vaultof given token
func getTTL(v *vault_dev.VaultDev, token string, i *instanceToken.InstanceToken) (ttl int, err error) {

	s, err := i.TokenLookup(token)
	if err != nil {
		return -1, err
	}

	if s == nil {
		return -1, fmt.Errorf("Error, no secret from init token lookup: %s", token)
	}

	dat, ok := s.Data["ttl"]
	if !ok {
		return -1, fmt.Errorf("Error ttl policy data from init token lookup")
	}
	// This is bad --
	str := fmt.Sprintf("%s", dat)
	// --

	n, err := strconv.Atoi(str)
	if err != nil {
		return -1, fmt.Errorf("%s", err)
	}

	return n, nil
}

// Init instace token for testing
func initInstanceToken(t *testing.T, vaultDev *vault_dev.VaultDev) *instanceToken.InstanceToken {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	log := logrus.NewEntry(logger)

	i := instanceToken.New(vaultDev.Client(), log)
	i.SetRole("master")
	i.SetClusterID("test-cluster")

	// setup temporary directory for tests
	dir, err := ioutil.TempDir("", "vault-helper-init-token")
	if err != nil {
		t.Fatal(err)
	}
	tempDirs = append(tempDirs, dir)
	i.SetVaultConfigPath(dir)

	if _, err := os.Stat(i.InitTokenFilePath()); os.IsNotExist(err) {
		ifile, err := os.Create(i.InitTokenFilePath())
		if err != nil {
			t.Fatalf("%s", err)
		}
		defer ifile.Close()
	}

	_, err = os.Stat(i.TokenFilePath())
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

// Check the token in file to be corrent
func tokenCheckFiles(t *testing.T, i *instanceToken.InstanceToken) {
	fileToken, err := i.TokenFromFile(i.TokenFilePath())
	if err != nil {
		t.Errorf("%s", err)
	}
	if fileToken != i.Token() {
		t.Fatalf("Token in file should equal the one that has been renewed. Exp=%s Got=%s", i.Token(), fileToken)
	}

	fileToken, err = i.TokenFromFile(i.InitTokenFilePath())
	if err != nil {
		t.Errorf("%s", err)
	}
	if fileToken != "" {
		t.Fatalf("Expexted no token in file '%s' but got= '%s'", i.InitTokenFilePath(), fileToken)
	}

	return
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
