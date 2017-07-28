package instanceToken

import (
	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

type InstanceToken struct {
	token       string
	role        string
	Log         *logrus.Entry
	clusterID   string
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
		i.Log = logger
	}

	return i

}
