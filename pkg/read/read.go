package read

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

type Read struct {
	vaultPath string
	fieldName string
	filePath  string
	owner     string
	group     string

	vaultClient *vault.Client
	Log         *logrus.Entry
}

func (r *Read) RunRead() error {
	//Read vault
	sec, err := r.vaultClient.Logical().Read(r.VaultPath())
	if err != nil {
		return fmt.Errorf("error reading from vault: %v", err)
	}

	if sec == nil {
		return errors.New("vault returned nothing")
	}

	var res string
	//Just get field
	if r.FieldName() != "" {
		res, err = r.getField(sec)
	} else {
		res, err = r.getPrettyJSON(sec)
	}
	if err != nil {
		return err
	}

	//Output to console
	if r.FilePath() == "" {
		str := ""
		if r.FieldName() != "" {
			str = "(" + r.FieldName() + ")"
		}
		str = "No file given. Outputting to console. " + str
		r.Log.Info(str)

		r.Log.Info(res)

		return nil
	}

	//Write to file
	r.Log.Infof("Outputing responce to file: %s", r.filePath)
	return r.writeToFile(res)
}

func (r *Read) getField(sec *vault.Secret) (field string, err error) {
	dat := sec.Data

	fieldDat, ok := dat[r.FieldName()]
	if !ok {
		return "", errors.New("error extracting field data from responce")
	}

	field, ok = fieldDat.(string)
	if !ok {
		b, ok := fieldDat.(bool)
		if !ok {
			i, ok := fieldDat.(json.Number)
			if !ok {
				return "", fmt.Errorf("error converting field data into string: %s", r.FieldName())
			}
			return string(i), nil
		}
		return strconv.FormatBool(b), nil
	}

	return field, nil
}

func (r *Read) writeToFile(res string) error {

	byt := []byte(res)
	if err := ioutil.WriteFile(r.FilePath(), byt, 0600); err != nil {
		return fmt.Errorf("error trying to write responce to file '%s': %s", r.FilePath(), err)
	}

	return r.writePermissons()
}

func (r *Read) writePermissons() error {

	if err := os.Chmod(r.FilePath(), os.FileMode(0600)); err != nil {
		return fmt.Errorf("error changing permissons of file '%s' to 0600: %v", r.FilePath(), err)
	}

	var uid int
	var gid int

	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("error getting current user info: %v", err)
	}

	if r.Owner() == "" {

		uid, err = strconv.Atoi(usr.Uid)
		if err != nil {
			return fmt.Errorf("error converting user uid '%s' (string) to (int): %v", usr.Uid, err)
		}

	} else {

		u, err := user.Lookup(r.Owner())
		if err != nil {
			return fmt.Errorf("error finding owner '%s' on system: %v", r.Owner(), err)
		}

		uid, err = strconv.Atoi(u.Uid)
		if err != nil {
			return fmt.Errorf("wrror converting user uid '%s' (string) to (int): %v", u.Uid, err)
		}

	}

	if r.Group() == "" {

		gid, err = strconv.Atoi(usr.Gid)
		if err != nil {
			return fmt.Errorf("error converting group gid '%s' (string) to (int): %v", usr.Gid, err)
		}

	} else {

		g, err := user.LookupGroup(r.Group())
		if err != nil {
			return fmt.Errorf("error finding group '%s' on system: %v", r.Group(), err)
		}

		gid, err = strconv.Atoi(g.Gid)
		if err != nil {
			return fmt.Errorf("error converting group gid '%s' (string) to (int): %v", g.Gid, err)
		}

	}

	if err := os.Chown(r.FilePath(), uid, gid); err != nil {
		return fmt.Errorf("error changing group and owner of file '%s' to usr:'%s' grp:'%s': %v", r.FilePath(), r.Owner(), r.Group(), err)
	}

	r.Log.Debugf("Set permissons on file: %s", r.FilePath())

	return nil
}

func (r *Read) getPrettyJSON(sec *vault.Secret) (prettyStr string, err error) {
	js, err := json.Marshal(sec)
	if err != nil {
		return "", fmt.Errorf("error converting responce from vault into JSON: %v", err)
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, js, "", "\t")
	if err != nil {
		return "", fmt.Errorf("error parsing JSON: %v", err)
	}

	return string(prettyJSON.Bytes()), nil
}

func New(v *vault.Client, log *logrus.Entry) *Read {
	r := &Read{
		vaultPath: "",
		fieldName: "",
		filePath:  "",
		owner:     "",
		group:     "",

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

func (r *Read) SetVaultPath(path string) {
	r.vaultPath = path
}
func (r *Read) VaultPath() (path string) {
	return r.vaultPath
}

func (r *Read) SetFieldName(name string) {
	r.fieldName = name
}
func (r *Read) FieldName() (name string) {
	return r.fieldName
}

func (r *Read) SetFilePath(path string) {
	r.filePath = path
}
func (r *Read) FilePath() (path string) {
	return r.filePath
}

func (r *Read) SetOwner(name string) {
	r.owner = name
}
func (r *Read) Owner() (name string) {
	return r.owner
}

func (r *Read) SetGroup(name string) {
	r.group = name
}
func (r *Read) Group() (name string) {
	return r.group
}
