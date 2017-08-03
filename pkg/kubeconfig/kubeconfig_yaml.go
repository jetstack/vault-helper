package kubeconfig

import (
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type KubeY struct {
	CurrentContext string
	ApiVersion     string
	Kind           string

	Clusters []Cluster
	Contexts []Context
	Users    []User
}

type Cluster struct {
	Name    string
	Cluster Clust
}
type Clust struct {
	Server                   string
	ApiVersion               string
	CertificateAuthorityData string
}

type Context struct {
	Name    string
	Context Conx
}
type Conx struct {
	Cluster   string
	Namespace string
	User      string
}

type User struct {
	Name string
	User Usr
}
type Usr struct {
	ClientCertificateData string
	ClientKeyData         string
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

func (u *Kubeconfig) StoreYaml(yml KubeY) error {

	marsh, err := yaml.Marshal(yml)
	if err != nil {
		return err
	}

	u.Log.Debugf("Created Yaml sucessfully.")

	fmt.Println(string(marsh))
	return nil
}

func (u *Kubeconfig) BuildYaml() (yml KubeY, err error) {

	clusterID := filepath.Base(u.cert.Role())
	apiURL := u.vaultClient.Address()

	cluster := Cluster{clusterID, Clust{"v1", apiURL, u.Cert64()}}
	context := Context{clusterID, Conx{clusterID, "kube-system", clusterID}}
	user := User{clusterID, Usr{u.Cert64(), u.CertKey64()}}

	yml = KubeY{
		CurrentContext: clusterID,
		ApiVersion:     "v1",
		Kind:           "Config",
		Clusters:       []Cluster{cluster},
		Contexts:       []Context{context},
		Users:          []User{user},
	}

	return yml, err
}

func (u *Kubeconfig) encode64File(path string) (byt string, err error) {
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("Expected file does not exist '%s':\n%s", path, err)
	} else if err != nil {
		return "", fmt.Errorf("Unexpected error reading file '%s':\n%s", path, err)
	}

	fi, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("Unexpected error reading file '%s':\n%s", path, err)
	}

	// need to convert file to []byte for encoding
	fileinfo, err := fi.Stat()
	if err != nil {
		return "", fmt.Errorf("Unable to get file info '%s':\n%s", path, err)
	}

	size := fileinfo.Size()
	bytes := make([]byte, size)

	// read file content into bytes
	buffer := bufio.NewReader(fi)
	_, err = buffer.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("Unable to read bytes from file '%s':\n%s", path, err)
	}

	str := b64.StdEncoding.EncodeToString(bytes)

	return str, nil
}
