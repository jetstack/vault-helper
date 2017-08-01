package instanceToken

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

func (i *InstanceToken) TokenFromFile(path string) (token string, err error) {
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

	exists, err := i.fileExists(i.TokenFilePath())
	if err != nil {
		return "", fmt.Errorf("Error checking file exists: %s", err)
	}

	if exists {
		i.Log.Debugf("File exists: %s", i.TokenFilePath())
		token, err := i.TokenFromFile(i.TokenFilePath())
		if err != nil {
			return "", fmt.Errorf("%s", err)
		}
		return token, nil
	}
	return "", nil
}

func (i *InstanceToken) WriteTokenFile(filePath, token string) error {

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if _, err := os.Create(filePath); err != nil {
			return fmt.Errorf("Error creating token file:\n%s", err)
		}
	}

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

func (i *InstanceToken) WipeTokenFile(filePath string) error {

	if err := deleteFile(filePath); err != nil {
		return fmt.Errorf("Error deleting token file '%s' to be wiped: \n%s", filePath, err)
	}

	if err := createFile(filePath); err != nil {
		return fmt.Errorf("Error creating token file '%s' that was wiped: \n%s", filePath, err)
	}

	return nil
}

func deleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

func createFile(path string) error {
	// detect if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		file, err := os.Create(path)

		if err != nil {
			return err
		}
		defer file.Close()
	}

	return nil
}

func (i *InstanceToken) initTokenNew() error {

	exists, err := i.fileExists(i.InitTokenFilePath())
	if err != nil {
		return fmt.Errorf("Error checking file exists: %s", err)
	}
	if !exists {
		return fmt.Errorf("No init token file: '%s' exiting.", i.InitTokenFilePath())
	}
	i.Log.Debugf("File exists at: %s", i.InitTokenFilePath())
	initToken, err := i.TokenFromFile(i.InitTokenFilePath())
	if err != nil {
		return fmt.Errorf("Error reading init token from file: %s", err)
	}
	if initToken == "" {
		return fmt.Errorf("Init token was not read from file: %s", i.InitTokenFilePath())
	}

	i.Log.Debugf("Init token found '%s' at '%s'", initToken, i.InitTokenFilePath())

	// Check init policies and init role are set (in enviroment?). Exit here if they are not.

	policies, err := i.TokenPolicies(initToken)
	if err != nil {
		return fmt.Errorf("Error finding init token policies: \n%s", err)
	}

	newToken, err := i.createToken(policies)
	if err != nil {
		return err
	}
	i.SetToken(newToken)

	i.Log.Infof("New token: %s", i.Token())

	return nil
}

func (i *InstanceToken) TokenPolicies(token string) (policies []string, err error) {

	s, err := i.TokenLookup(token)
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
		DisplayName: i.Role(),
		Policies:    policies,
	}

	newToken, err := i.vaultClient.Auth().Token().CreateOrphan(tCreateRequest)
	if err != nil {
		return "", fmt.Errorf("failed to create init token: %s", err)
	}

	return newToken.Auth.ClientToken, nil
}

func (i *InstanceToken) TokenLookup(token string) (secret *vault.Secret, err error) {
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

	s, err := i.TokenLookup(i.Token())
	if err != nil {
		return nil
	}

	dat, ok := s.Data["renewable"]
	if !ok {
		return fmt.Errorf("Unable to get renewable token data from secret")
	}

	if dat == false {
		return fmt.Errorf("Token not renewable: %s", i.Token())
	}
	i.Log.Debugf("Token renewable")

	// Renew against vault
	s, err = i.vaultClient.Auth().Token().Renew(i.Token(), 0)
	if err != nil {
		return fmt.Errorf("Error renewing token %s: %s - %s", i.Role(), i.Token(), err)
	}

	i.Log.Infof("Renewed token: %s", i.Token())

	return nil
}

func (i *InstanceToken) TokenRenewRun() error {

	token, err := i.TokenRetrieve()
	if err != nil && os.IsExist(err) {
		return fmt.Errorf("Error retreiving token from file: %s", err)
	}

	if token != "" {
		// Token exists in file
		// Renew token
		logrus.Debugf("Token to renew: %s", token)
		i.SetToken(token)
		if err := i.tokenRenew(); err != nil {
			return err
		}
		return nil
	}

	//Token Doesn't exist
	i.Log.Infof("Token doesn't exist, generating new")
	err = i.initTokenNew()
	if err != nil {
		return fmt.Errorf("Error generating new token: \n%s", err)
	}

	if err := i.WriteTokenFile(i.TokenFilePath(), i.Token()); err != nil {
		return fmt.Errorf("Error writting token to file: %s", err)
	}
	if err := i.WipeTokenFile(i.InitTokenFilePath()); err != nil {
		return fmt.Errorf("Error wiping token from file: %s", err)
	}

	i.Log.Infof("Token written to file: %s", i.TokenFilePath())

	return nil
}
