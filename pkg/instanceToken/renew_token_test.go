package instanceToken_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/jetstack-experimental/vault-helper/pkg/instanceToken"
	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"
	"github.com/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

const Token_File = "/etc/vault/token"
const Init_Token_File = "/etc/vault/init-token"

func TestMain(t *testing.T) {
	v := initVaultDev(t)
	initKubernetes(t, v)

	testRenew_Token_Exists(t, v)
	testRenew_Token_NotExists(t, v)
	testRenew_Token_NeitherExist(t, v)
	testRenew_Token_Exists_NoRenew(t, v)

	v.Stop()
}

// Token exists at token_file - renew
func testRenew_Token_Exists(t *testing.T, vaultDev *vault_dev.VaultDev) {

	k := initKubernetes(t, vaultDev)
	i := initInstanceToken(vaultDev)

	token := k.InitTokens()["master"]
	if err := i.WriteTokenFile(Token_File, token); err != nil {
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
func testRenew_Token_NotExists(t *testing.T, vaultDev *vault_dev.VaultDev) {

	k := initKubernetes(t, vaultDev)
	i := initInstanceToken(vaultDev)

	if err := i.WriteTokenFile(Init_Token_File, k.InitTokens()["master"]); err != nil {
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
func testRenew_Token_Exists_NoRenew(t *testing.T, vaultDev *vault_dev.VaultDev) {

	initKubernetes(t, vaultDev)
	i := initInstanceToken(vaultDev)

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

	if err := i.WriteTokenFile(Token_File, newToken.Auth.ClientToken); err != nil {
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
func testRenew_Token_NeitherExist(t *testing.T, vaultDev *vault_dev.VaultDev) {

	initKubernetes(t, vaultDev)
	i := initInstanceToken(vaultDev)

	err := i.TokenRenewRun()

	if err == nil {
		t.Fatalf("Expected an error - init file is empty")
	}

	i.Log.Debugf("%s", err)
	if err.Error() == "Error generating new token: \nInit token was not read from file: /etc/vault/init-token" {
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
func initInstanceToken(vaultDev *vault_dev.VaultDev) *instanceToken.InstanceToken {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	log := logrus.NewEntry(logger)

	i := instanceToken.New(vaultDev.Client(), log)
	i.SetRole("master")
	i.SetClusterID("test-cluster")

	if _, err := os.Stat("/etc/vault"); os.IsNotExist(err) {
		os.Mkdir("/etc/vault", os.ModeDir)
	}

	var _, err = os.Stat(Init_Token_File)
	if os.IsNotExist(err) {
		ifile, err := os.Create(Init_Token_File)
		if err != nil {
			logrus.Fatalf("%s", err)
		}
		defer ifile.Close()
	}

	_, err = os.Stat(Token_File)
	if os.IsNotExist(err) {
		tfile, err := os.Create(Token_File)
		if err != nil {
			logrus.Fatalf("%s", err)
		}
		defer tfile.Close()
	}

	i.WipeTokenFile(Init_Token_File)
	i.WipeTokenFile(Token_File)

	return i
}

// Check the token in file to be corrent
func tokenCheckFiles(t *testing.T, i *instanceToken.InstanceToken) {
	fileToken, err := i.TokenFromFile(Token_File)
	if err != nil {
		t.Errorf("%s", err)
	}
	if fileToken != i.Token() {
		t.Fatalf("Token in file should equal the one that has been renewed. Exp=%s Got=%s", i.Token(), fileToken)
	}

	fileToken, err = i.TokenFromFile(Init_Token_File)
	if err != nil {
		t.Errorf("%s", err)
	}
	if fileToken != "" {
		t.Fatalf("Expexted no token in file '%s' but got= '%s'", Init_Token_File, fileToken)
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
func initVaultDev(t *testing.T) *vault_dev.VaultDev {
	vaultDev := vault_dev.New()

	if err := vaultDev.Start(); err != nil {
		t.Fatalf("unable to initialise vault dev server for integration tests: %s", err)
	}

	return vaultDev
}
