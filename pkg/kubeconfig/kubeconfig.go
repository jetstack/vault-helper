package kubeconfig

import (
	//"encoding/pem"
	//"fmt"
	//"os"
	//"os/user"
	//"strconv"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/jetstack-experimental/vault-helper/pkg/cert"
)

type Kubeconfig struct {
	filePath string

	cert        *cert.Cert
	Log         *logrus.Entry
	vaultClient *vault.Client
}

func New(vaultClient *vault.Client, logger *logrus.Entry) *Kubeconfig {

	u := &Kubeconfig{}

	if vaultClient != nil {
		u.vaultClient = vaultClient
	}
	if logger != nil {
		u.Log = logger
	}

	return u
}

func (u *Kubeconfig) RunKube() error {

	return nil
}

func (u *Kubeconfig) SetCert(cert *cert.Cert) {
	u.cert = cert
}
func (u *Kubeconfig) Cert() (c *cert.Cert) {
	return u.cert
}

func (u *Kubeconfig) SetFilePath(path string) {
	u.filePath = path
}
func (u *Kubeconfig) FilePath() (path string) {
	return u.filePath
}
