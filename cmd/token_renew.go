package cmd

import (
	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/jetstack-experimental/vault-helper/pkg/instanceToken"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var renewtokenCmd = &cobra.Command{
	Use:   "renew-token [cluster ID]",
	Short: "Renew token on vault server",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()
		logger.Level = logrus.DebugLevel
		log := logrus.NewEntry(logger)

		v, err := vault.NewClient(nil)
		if err != nil {
			logger.Fatal(err)
		}

		i := instanceToken.New(v, log)

		if err := i.Run(cmd, args); err != nil {
			logger.Fatal(err)
		}

	},
}

func init() {

	renewtokenCmd.PersistentFlags().String(instanceToken.FlagTokenRole, "", "Set role of token to renew (default *no role*)")
	//renewtokenCmd.Flags().String(instanceToken.FlagVaultConfigPath, "/etc/vault", "Set config path to directory with tokens")

	RootCmd.AddCommand(renewtokenCmd)
}
