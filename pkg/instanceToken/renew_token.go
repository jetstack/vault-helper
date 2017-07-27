package instanceToken

import (
	"fmt"
	"io/ioutil"
	"os"

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

// NEw
// INititalise logrus.Entry

func New(vaultClient *vault.Client) *InstanceToken {
	i := &InstanceToken{
		role:      "",
		token:     "",
		clusterID: "",
	}

	if vaultClient != nil {
		i.vaultClient = vaultClient
	}

	return i

}

func tokenFromFile(path string) (token string, err error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	token = string(dat)

	return token, nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil

}

func TokenRetrieve() (token string, err error) {

	exists, err := fileExists(Token_File)
	if err != nil {
		return "", fmt.Errorf("Error checking file exists: %s", err)
	}

	if exists {
		logrus.Debugf("File exists: %s", Token_File)
		token, err := tokenFromFile(Token_File)
		if err != nil {
			return "", fmt.Errorf("%s", err)
		}
		return token, nil
	}
	return "", nil
}

func token_new() error {

	exists, err := fileExists(Init_Token_File)
	if err != nil {
		return fmt.Errorf("Error checking file exists: %s", err)
	}
	if !exists {
		return fmt.Errorf("No init token file: '%s' exiting.", Init_Token_File)
	}
	logrus.Debugf("File exists at: %s", Init_Token_File)
	init_token, err := tokenFromFile(Init_Token_File)
	if err != nil {
		return fmt.Errorf("Error reading init token from file: %s", err)
	}
	if init_token == "" {
		return fmt.Errorf("Init token was not read from file: %s", Init_Token_File)
	}
	logrus.Debugf("Init token found '%s' at '%s'", init_token, Init_Token_File)

	// Check init policies and init role are set (in enviroment?). Exit here if they are not.

	return nil

}

func token_renew(token string) error {
	// Check if renewable

	// Renew against vault

	return nil
}

func token_renew_run() {

	logrus.SetLevel(logrus.DebugLevel)

	token, err := TokenRetrieve()
	if err != nil {
		logrus.Errorf("Error retreiving token from file: %s", err)
	}

	if token != "" {
		// Token exists in file
		// Renew token
	}

	//Token Doesn't exist
	logrus.Debugf("Token doesn't exist, generating new")
	err = token_new()
	if err != nil {
		logrus.Errorf("Error generating new token: %s", err)
	}

}
