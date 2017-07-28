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

func (i *InstanceToken) SetClusterID(custer string) {

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
