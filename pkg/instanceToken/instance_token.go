package instanceToken

import (
	"path/filepath"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

type InstanceToken struct {
	token           string
	role            string
	clusterID       string
	vaultConfigPath string

	Log         *logrus.Entry
	vaultClient *vault.Client
}

func (i *InstanceToken) SetRole(role string) {
	i.role = role
}

func (i *InstanceToken) Role() (role string) {
	return i.role
}

func (i *InstanceToken) SetToken(token string) {
	i.token = token
}

func (i *InstanceToken) Token() (token string) {
	return i.token
}

func (i *InstanceToken) SetClusterID(clusterID string) {
	i.clusterID = clusterID
}

func (i *InstanceToken) SetVaultConfigPath(path string) {
	i.vaultConfigPath = path
}

func (i *InstanceToken) VaultConfigPath() (path string) {
	return i.vaultConfigPath
}

func (i *InstanceToken) TokenFilePath() (path string) {
	return filepath.Join(i.VaultConfigPath(), "token")
}
func (i *InstanceToken) InitTokenFilePath() (path string) {
	return filepath.Join(i.VaultConfigPath(), "init-token")
}

func New(vaultClient *vault.Client, logger *logrus.Entry) *InstanceToken {
	i := &InstanceToken{
		role:            "",
		token:           "",
		clusterID:       "",
		vaultConfigPath: "/etc/vault",
	}

	if vaultClient != nil {
		i.vaultClient = vaultClient
	}

	if logger != nil {
		i.Log = logger
	}

	return i

}
