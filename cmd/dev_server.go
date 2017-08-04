package cmd

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jetstack-experimental/vault-helper/pkg/dev_server"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var devServerCmd = &cobra.Command{
	Use:   "dev-server [cluster ID]",
	Short: "Run a vault server in development mode with kubernetes PKI created",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			logrus.Fatalf("no cluster ID was given")
		}

		v := dev_server.New()

		if err := v.Run(cmd, args); err != nil {
			logrus.Fatal(err)
		}

		for n, t := range v.Kubernetes.InitTokens() {
			logrus.Infof(n + "-init_token := " + t)
		}

	},
}

func init() {
	devServerCmd.PersistentFlags().Duration(dev_server.FlagMaxValidityCA, time.Hour*24*365*20, "Maxium validity for CA certificates")
	devServerCmd.PersistentFlags().Duration(dev_server.FlagMaxValidityAdmin, time.Hour*24*365, "Maxium validity for admin certificates")
	devServerCmd.PersistentFlags().Duration(dev_server.FlagMaxValidityComponents, time.Hour*24*30, "Maxium validity for component certificates")

	devServerCmd.PersistentFlags().String(dev_server.FlagInitTokenEtcd, "", "Set init-token-etcd   (Default to new token)")
	devServerCmd.PersistentFlags().String(dev_server.FlagInitTokenWorker, "", "Set init-token-worker (Default to new token)")
	devServerCmd.PersistentFlags().String(dev_server.FlagInitTokenMaster, "", "Set init-token-master (Default to new token)")
	devServerCmd.PersistentFlags().String(dev_server.FlagInitTokenAll, "", "Set init-token-all    (Default to new token)")

	RootCmd.AddCommand(devServerCmd)
}
