package read

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

type Read struct {
	clusterID string
	vaultPath string
	filePath  string

	vaultClient *vault.Client
	Log         *logrus.Entry
}

func (r *Read) RunRead() error {

	//Read vault
	sec, err := r.vaultClient.Logical().Read(r.VaultPath())
	if err != nil {
		return fmt.Errorf("Error reading from vault:\n%s", err)
	}

	if sec == nil {
		r.Log.Warnf("Vault returned nothing.")
		return nil
	}

	res, err := r.getPrettyJSON(sec)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	//Output to console
	if r.filePath == "" {
		r.Log.Infof("No file given. Outputting to console.")
		r.Log.Infof("%s", res)
		return nil
	}

	//Write to file
	r.Log.Infof("Outputing responce to file: %s", r.filePath)
	return r.writeToFile(res)
}

func (r *Read) writeToFile(res string) error {

	byt := []byte(res)
	if err := ioutil.WriteFile(r.FilePath(), byt, 0600); err != nil {
		return fmt.Errorf("Error trying to write responce to file '%s':\n%s", r.FilePath(), err)
	}

	return nil
}

func (r *Read) getPrettyJSON(sec *vault.Secret) (prettyStr string, err error) {

	js, err := json.Marshal(sec)
	if err != nil {
		return "", fmt.Errorf("Error converting responce from vault into JSON:\n%s", err)
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, js, "", "\t")
	if err != nil {
		return "", fmt.Errorf("Error parsing JSON:\n%s", err)
	}

	return string(prettyJSON.Bytes()), nil
}

func New(v *vault.Client, log *logrus.Entry) *Read {
	r := &Read{
		clusterID: "",
		vaultPath: "",
		filePath:  "",

		vaultClient: v,
		Log:         log,
	}

	if v != nil {
		r.vaultClient = v
	}

	if log != nil {
		r.Log = log
	}

	return r
}

func (r *Read) SetClusterID(clusterID string) {
	r.clusterID = clusterID
}
func (r *Read) ClusterID() (clusterID string) {
	return r.clusterID
}

func (r *Read) SetVaultPath(path string) {
	r.vaultPath = path
}
func (r *Read) VaultPath() (path string) {
	return r.vaultPath
}

func (r *Read) SetFilePath(path string) {
	r.filePath = path
}
func (r *Read) FilePath() (path string) {
	return r.filePath
}
