package cmd

import (
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	"github.com/jetstack-experimental/vault-helper/pkg/instanceToken"
)

// initCmd represents the init command
var renewtokenCmd = &cobra.Command{
	Use:   "renew-token [cluster ID]",
	Short: "Renew token on vault server.",
	Run: func(cmd *cobra.Command, args []string) {
		log := LogLevel(cmd)

		v, err := vault.NewClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		i := instanceToken.New(v, log)

		if err := i.Run(cmd, args); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	renewtokenCmd.PersistentFlags().String(instanceToken.FlagTokenRole, "", "Set role of token to renew. (default *no role*)")

	RootCmd.AddCommand(renewtokenCmd)
}
