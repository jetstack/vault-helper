package cmd

import (
	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/jetstack-experimental/vault-helper/pkg/instanceToken"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var renewtokenCmd = &cobra.Command{
	Use:   "renew-token",
	Short: "Renew token on vault server",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()
		logger.Level = logrus.DebugLevel
		//logger.WithField(

		v, err := vault.NewClient(nil)
		if err != nil {
			logger.Fatal(err)
		}

		i := instanceToken.New(v)

		if err := i.Run(cmd, args); err != nil {
			logger.Fatal(err)
		}

	},
}

func init() {
	renewtokenCmd.PersistentFlags().String(instanceToken.FlagTokenRole, "", "Set role of token to renew (Default to *no role*")

	RootCmd.AddCommand(renewtokenCmd)
}
