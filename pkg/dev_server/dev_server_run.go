package dev_server

import (
	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"

	"github.com/spf13/cobra"
)

const FlagMaxValidityAdmin = "max-validity-admin"
const FlagMaxValidityCA = "max-validity-ca"
const FlagMaxValidityComponents = "max-validity-components"

const FlagInitTokenEtcd = "init-token-etcd"
const FlagInitTokenAll = "init-token-all"
const FlagInitTokenMaster = "init-token-master"
const FlagInitTokenWorker = "init-token-worker"

const FlagWaitSignal = "wait-signal"
const FlagPortNumber = "port"

func (v *DevVault) Run(cmd *cobra.Command, args []string) error {

	v.Kubernetes = kubernetes.New(v.Vault.Client(), v.Log)

	return v.Kubernetes.Run(cmd, args)
}
