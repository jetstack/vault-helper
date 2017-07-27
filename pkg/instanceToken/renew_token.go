package instanceToken

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

const Token_File = "/etc/vault/token"
const Init_Token_File = "/etc/vault/init-token"

type InstanceToken struct {
	token       string
	role        string
	log         *logrus.Entry
	clusterID   string
	vaultClient *vault.Client
}

func New(vaultClient *vault.Client, logger *logrus.Entry) *InstanceToken {
	i := &InstanceToken{
		role:      "",
		token:     "",
		clusterID: "",
	}

	if vaultClient != nil {
		i.vaultClient = vaultClient
	}

	if logger != nil {
		i.log = logger
	}

	return i

}

func (i *InstanceToken) tokenFromFile(path string) (token string, err error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	str := string(dat)
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\n", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	token = strings.Replace(str, "\r", "", -1)
	//token = strings.Replace(str, "-", "", -1) //I beleive these may be needed

	return token, nil
}

func (i *InstanceToken) fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil

}

func (i *InstanceToken) TokenRetrieve() (token string, err error) {

	exists, err := i.fileExists(Token_File)
	if err != nil {
		return "", fmt.Errorf("Error checking file exists: %s", err)
	}

	if exists {
		i.log.Debugf("File exists: %s", Token_File)
		token, err := i.tokenFromFile(Token_File)
		if err != nil {
			return "", fmt.Errorf("%s", err)
		}
		return token, nil
	}
	return "", nil
}

func (i *InstanceToken) tokenNew() error {

	exists, err := i.fileExists(Init_Token_File)
	if err != nil {
		return fmt.Errorf("Error checking file exists: %s", err)
	}
	if !exists {
		return fmt.Errorf("No init token file: '%s' exiting.", Init_Token_File)
	}
	i.log.Debugf("File exists at: %s", Init_Token_File)
	init_token, err := i.tokenFromFile(Init_Token_File)
	if err != nil {
		return fmt.Errorf("Error reading init token from file: %s", err)
	}
	if init_token == "" {
		return fmt.Errorf("Init token was not read from file: %s", Init_Token_File)
	}

	i.log.Debugf("Init token found '%s' at '%s'", init_token, Init_Token_File)

	// Check init policies and init role are set (in enviroment?). Exit here if they are not.

	return nil

}

func (i *InstanceToken) tokenRenew() error {
	// Check if renewable
	s, err := i.vaultClient.Auth().Token().Lookup(i.token)
	if err != nil {
		return fmt.Errorf("Error looking up token %s: %s - %s", i.role, i.token, err)
	}

	if s == nil {
		return fmt.Errorf("Error finding secret with token from vault '%s'", i.token)
	}

	dat, ok := s.Data["renewable"]
	if !ok {
		return fmt.Errorf("Unable to get renewable token data from secret")
	}
	if dat == false {
		i.log.Infof("Token not renewable")
		return nil
	}
	i.log.Debugf("Token renewable")

	// Renew against vault
	s, err = i.vaultClient.Auth().Token().Renew(i.token, 0)
	if err != nil {
		return fmt.Errorf("Error renewing token %s: %s - %s", i.role, i.token, err)
	}

	i.log.Debugf("Renewed token: %s", i.token)

	return nil
}

func (i *InstanceToken) tokenRenewRun() error {

	token, err := i.TokenRetrieve()
	if err != nil {
		i.log.Errorf("Error retreiving token from file: %s", err)
	}

	if token != "" {
		// Token exists in file
		// Renew token
		logrus.Debugf("Token to renew: %s", token)
		i.token = token
		if err := i.tokenRenew(); err != nil {
			return err
		}
		return nil
	}

	//Token Doesn't exist
	i.log.Debugf("Token doesn't exist, generating new")
	err = i.tokenNew()
	if err != nil {
		i.log.Errorf("Error generating new token: %s", err)
	}

	return nil
}
