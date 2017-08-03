package dev_server

import (
	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"
	"github.com/spf13/cobra"
)

const FlagMaxValidityAdmin = "max-validity-admin"
const FlagMaxValidityCA = "max-validity-ca"
const FlagMaxValidityComponents = "max-validity-components"

const FlagInitToken_etcd = "init-token-etcd"
const FlagInitToken_all = "init-token-all"
const FlagInitToken_master = "init-token-master"
const FlagInitToken_worker = "init-token-worker"

func (v *DevVault) Run(cmd *cobra.Command, args []string) error {

	k := kubernetes.New(v.Vault.Client())
	v.Kubernetes = k

	return k.Run(cmd, args)
}
