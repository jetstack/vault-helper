package cmd

import (
	//"github.com/Sirupsen/logrus"
	//vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	//"fmt"
	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/kubernetes"
	//"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

// initCmd represents the init command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup kubernetes on a running vault server",
	Run: func(cmd *cobra.Command, args []string) {

		kubernetes.Run(cmd, args)

	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
