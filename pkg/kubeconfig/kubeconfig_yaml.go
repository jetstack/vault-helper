package kubeconfig

import (
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type KubeY struct {
	CurrentContext string `yaml:"current-context"`
	ApiVersion     string `yaml:"apiVersion"`
	Kind           string `yaml:"kind"`

	Clusters []Cluster
	Contexts []Context
	Users    []User
}

type Cluster struct {
	Name    string `yaml:"name"`
	Cluster Clust
}
type Clust struct {
	Server                   string `yaml:"server"`
	ApiVersion               string `yaml:"api-version"`
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
}

type Context struct {
	Name    string `yaml:"name"`
	Context Conx
}
type Conx struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace"`
	User      string `yaml:"user"`
}

type User struct {
	Name string `yaml:"name"`
	User Usr
}
type Usr struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}

func (u *Kubeconfig) EncodeCerts() error {

	byt, err := u.encode64File(u.Cert().Destination() + "-key.pem")
	if err != nil {
		return err
	}
	u.SetCertKey64(byt)

	byt, err = u.encode64File(u.Cert().Destination() + "-ca.pem")
	if err != nil {
		return err
	}
	u.SetCertCA64(byt)

	byt, err = u.encode64File(u.Cert().Destination() + ".pem")
	if err != nil {
		return err
	}
	u.SetCert64(byt)

	return nil
}

func (u *Kubeconfig) StoreYaml(yml string) error {
	path := filepath.Clean(u.FilePath())

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating yaml file at '%s': %s", path, err)
	}
	defer file.Close()

	if _, err := file.Write([]byte(yml)); err != nil {
		return fmt.Errorf("error writting to yaml file '%s': %s", path, err)
	}

	u.Log.Infof("Yaml writting to file: %s", path)

	return u.WritePermissions()
}

func (u *Kubeconfig) WritePermissions() error {
	if err := os.Chmod(u.FilePath(), os.FileMode(0600)); err != nil {
		return fmt.Errorf("error changing permissons of file '%s' to 0600:%s", u.FilePath(), err)
	}

	var uid int
	var gid int

	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user info:%s", err)
	}

	if u.Cert().Owner() == "" {

		uid, err = strconv.Atoi(usr.Uid)
		if err != nil {
			return fmt.Errorf("error converting user uid '%s' (string) to (int): %s", usr.Uid, err)
		}

	} else {

		us, err := user.Lookup(u.Cert().Owner())
		if err != nil {
			return fmt.Errorf("error finding owner '%s' on system: %s", u.Cert().Owner(), err)
		}

		uid, err = strconv.Atoi(us.Uid)
		if err != nil {
			return fmt.Errorf("error converting user uid '%s' (string) to (int): %s", us.Uid, err)
		}

	}

	if u.Cert().Group() == "" {

		gid, err = strconv.Atoi(usr.Gid)
		if err != nil {
			return fmt.Errorf("error converting group gid '%s' (string) to (int): %s", usr.Gid, err)
		}

	} else {

		g, err := user.LookupGroup(u.Cert().Group())
		if err != nil {
			return fmt.Errorf("error finding group '%s' on system: %s", u.Cert().Group(), err)
		}

		gid, err = strconv.Atoi(g.Gid)
		if err != nil {
			return fmt.Errorf("error converting group gid '%s' (string) to (int): %s", g.Gid, err)
		}

	}

	if err := os.Chown(u.FilePath(), uid, gid); err != nil {
		return fmt.Errorf("error changing group and owner of file '%s' to usr:'%s' grp:'%s': %s", u.FilePath(), u.Cert().Owner(), u.Cert().Group(), err)
	}

	u.Log.Debugf("Set permissons on file: %s", u.FilePath())

	return nil

}

func (u *Kubeconfig) BuildYaml() (yml string, err error) {

	path := filepath.Clean(u.cert.Role())
	clusterID := strings.Split(path, "/")[0]
	apiURL := u.vaultClient.Address()

	cluster := Cluster{clusterID, Clust{apiURL, "v1", u.CertCA64()}}
	context := Context{clusterID, Conx{clusterID, "kube-system", clusterID}}
	user := User{clusterID, Usr{u.Cert64(), u.CertKey64()}}

	ky := KubeY{
		CurrentContext: clusterID,
		ApiVersion:     "v1",
		Kind:           "Config",
		Clusters:       []Cluster{cluster},
		Contexts:       []Context{context},
		Users:          []User{user},
	}

	marsh, err := yaml.Marshal(ky)
	if err != nil {
		return "", err
	}
	u.Log.Debugf("Created Yaml sucessfully.")

	return string(marsh), err
}

func (u *Kubeconfig) encode64File(path string) (byt string, err error) {
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("expected file does not exist '%s': %s", path, err)
	} else if err != nil {
		return "", fmt.Errorf("unexpected error reading file '%s': %s", path, err)
	}

	fi, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("unexpected error reading file '%s': %s", path, err)
	}

	// need to convert file to []byte for encoding
	fileinfo, err := fi.Stat()
	if err != nil {
		return "", fmt.Errorf("unable to get file info '%s': %s", path, err)
	}

	size := fileinfo.Size()
	bytes := make([]byte, size)

	// read file content into bytes
	buffer := bufio.NewReader(fi)
	_, err = buffer.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("unable to read bytes from file '%s': %s", path, err)
	}

	str := b64.StdEncoding.EncodeToString(bytes)

	return str, nil
}
