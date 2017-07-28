package instanceToken_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jetstack-experimental/vault-helper/pkg/instanceToken"
	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"
	"github.com/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

var vault *vault_dev.VaultDev
var test int

const tests = 2
const Token_File = "/etc/vault/token"
const Init_Token_File = "/etc/vault/init-token"

func TestRenew_Token_Exists(t *testing.T) {
	test++

	if vault == nil {
		initVaultDev(t)
	}
	if test == tests {
		defer vault.Stop()
	}

	k := initKubernetes(t)
	i := initInstanceToken()

	token := k.InitTokens()["master"]
	if err := i.WriteTokenFile(Token_File, token); err != nil {
		t.Errorf("Error setting token for test: \n%s", err)
		return
	}

	i.Log.Debugf("Waiting for lower ttl . . .")
	time.Sleep(time.Second * 5)

	ttl, err := getTTL(vault, token, i)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	if err := i.TokenRenewRun(); err != nil {
		t.Errorf("Error renewing token from token file (Exists): \n%s", err)
		return
	}

	newttl, err := getTTL(vault, token, i)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	i.Log.Debugf("old ttl: %ss    new ttl: %ss", strconv.Itoa(ttl), strconv.Itoa(newttl))

	if ttl > newttl {
		t.Errorf("Token was not renewed - old ttl higher than new\nold=%s new=%s", strconv.Itoa(ttl), strconv.Itoa(newttl))
		return
	}

	fileToken, err := i.TokenFromFile(Token_File)
	if err != nil {
		t.Errorf("%s", err)
	}
	if fileToken != token {
		t.Errorf("Token in file should equal the one that has been renewed. Exp=%s Got=%s", token, fileToken)
		return
	}

	fileToken, err = i.TokenFromFile(Init_Token_File)
	if err != nil {
		t.Errorf("%s", err)
	}
	if fileToken != "" {
		t.Errorf("Expexted no token in file '%s' but got= '%s'", Init_Token_File, fileToken)
		return
	}

}

func TestRenew_Token_NotExists(t *testing.T) {
	test++

	if vault == nil {
		initVaultDev(t)
	}
	if test == tests {
		defer vault.Stop()
	}

	k := initKubernetes(t)
	i := initInstanceToken()

	token := k.InitTokens()["master"]
	if err := i.WriteTokenFile(Token_File, token); err != nil {
		t.Errorf("Error setting token for test: \n%s", err)
		return
	}
}

// TODO: Token doesn't exist at /etc/vault/token, create token, renew
// TODO: Token exists but can't be renewed
// TODO: Token doesn't exist at either file

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

func initInstanceToken() *instanceToken.InstanceToken {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	log := logrus.NewEntry(logger)

	i := instanceToken.New(vault.Client(), log)
	i.SetRole("master")

	if _, err := os.Stat("/etc/vault"); os.IsNotExist(err) {
		os.Mkdir("/etc/vault", os.ModeDir)
	}

	var _, err = os.Stat(Init_Token_File)
	if os.IsNotExist(err) {
		ifile, err := os.Create(Init_Token_File)
		if err != nil {
			logrus.Errorf("%s", err)
			return nil
		}
		defer ifile.Close()
	}

	_, err = os.Stat(Token_File)
	if os.IsNotExist(err) {
		tfile, err := os.Create(Token_File)
		if err != nil {
			logrus.Errorf("%s", err)
			return nil
		}
		defer tfile.Close()
	}

	i.WipeTokenFile(Init_Token_File)
	i.WipeTokenFile(Token_File)

	return i
}

func initKubernetes(t *testing.T) *kubernetes.Kubernetes {
	k := kubernetes.New(vault.Client())
	k.SetClusterID("test-cluster")

	if err := k.Ensure(); err != nil {
		t.Errorf("Error ensuring kubernetes: \n%s", err)
		return nil
	}

	return k
}

func initVaultDev(t *testing.T) {
	vault = vault_dev.New()

	if err := vault.Start(); err != nil {
		t.Skip("unable to initialise vault dev server for integration tests: ", err)
		return
	}
}
