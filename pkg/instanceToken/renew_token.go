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

func (i *InstanceToken) writeTokenFile(filePath, token string) error {

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("Error opening file: %s\n %s", filePath, err)
	}

	defer f.Close()

	if _, err = f.WriteString(token); err != nil {
		return fmt.Errorf("Error writting to file: %s\n %s", filePath, err)
	}
	return nil
}

func (i *InstanceToken) wipeTokenFile(filePath string) error {

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("Error opening file: %s\n %s", filePath, err)
	}

	defer f.Close()

	if _, err = f.WriteString(""); err != nil {
		return fmt.Errorf("Error writting to file: %s\n %s", filePath, err)
	}
	return nil
}

func (i *InstanceToken) initTokenNew() error {

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

	policies, err := i.tokenPolicies(init_token)
	if err != nil {
		return fmt.Errorf("Error finding init token policies: \n%s", err)
	}

	newToken, err := i.createToken(policies)
	if err != nil {
		return err
	}
	i.token = newToken

	i.log.Debugf("New token: %s", i.token)

	return nil
}

func (i *InstanceToken) tokenPolicies(token string) (policies []string, err error) {

	s, err := i.tokenLookup(token)
	if err != nil {
		return nil, err
	}

	if s == nil {
		return nil, fmt.Errorf("Error, no secret from init token lookup: %s", token)
	}

	dat, ok := s.Data["policies"]
	if !ok {
		return nil, fmt.Errorf("Error getting policy data from init token lookup")
	}

	d, ok := dat.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Error converting data to []interface")
	}

	policies = make([]string, len(d))

	for n, m := range d {
		str, ok := m.(string)
		if !ok {
			return nil, fmt.Errorf("Error converting interface to string")
		}
		policies[n] = str
	}

	return policies, nil
}

func (i *InstanceToken) createToken(policies []string) (token string, err error) {

	tCreateRequest := &vault.TokenCreateRequest{
		DisplayName: i.role,
		Policies:    policies,
	}

	newToken, err := i.vaultClient.Auth().Token().CreateOrphan(tCreateRequest)
	if err != nil {
		return "", fmt.Errorf("failed to create init token: %s", err)
	}

	return newToken.Auth.ClientToken, nil
}

func (i *InstanceToken) tokenLookup(token string) (secret *vault.Secret, err error) {
	s, err := i.vaultClient.Auth().Token().Lookup(token)
	if err != nil {
		return nil, fmt.Errorf("Error looking up token: %s\n \n%s", token, err)
	}

	if s == nil {
		return nil, fmt.Errorf("Error finding secret with token from vault '%s'", token)
	}

	return s, nil

}

func (i *InstanceToken) tokenRenew() error {
	// Check if renewable

	s, err := i.tokenLookup(i.token)
	if err != nil {
		return nil
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
		return fmt.Errorf("Error retreiving token from file: %s", err)
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
	err = i.initTokenNew()
	if err != nil {
		i.log.Errorf("Error generating new token: \n%s", err)
	}

	if err := i.writeTokenFile("/etc/vault/token", i.token); err != nil {
		return fmt.Errorf("Error writting token to file: %s", err)
	}
	if err := i.wipeTokenFile("/etc/vault/init-token"); err != nil {
		return fmt.Errorf("Error wiping token from file: %s", err)
	}

	i.log.Debugf("New init token written to file")

	return nil
}
