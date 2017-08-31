package cmd

import (
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	"github.com/jetstack-experimental/vault-helper/pkg/instanceToken"
	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"
)

// initCmd represents the init command
var setupCmd = &cobra.Command{
	Use:   "setup [cluster ID]",
	Short: "Setup kubernetes on a running vault server.",
	Run: func(cmd *cobra.Command, args []string) {
		log := LogLevel(cmd)

		v, err := vault.NewClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		instanceToken.SetVaultToken(v, log, cmd)

		k := kubernetes.New(v, log)
		if err != nil {
			log.Fatal(err)
		}

		if err := k.Run(cmd, args); err != nil {
			log.Fatal(err)
		}

		for n, t := range k.InitTokens() {
			log.Infof(n + "-init_token := " + t)
		}

	},
}

func init() {
	setupCmd.PersistentFlags().Duration(kubernetes.FlagMaxValidityCA, time.Hour*24*365*20, "Maxium validity for CA certificates")
	setupCmd.PersistentFlags().Duration(kubernetes.FlagMaxValidityAdmin, time.Hour*24*365, "Maxium validity for admin certificates")
	setupCmd.PersistentFlags().Duration(kubernetes.FlagMaxValidityComponents, time.Hour*24*30, "Maxium validity for component certificates")

	setupCmd.PersistentFlags().String(kubernetes.FlagInitToken_etcd, "", "Set init-token-etcd   (Default to new token)")
	setupCmd.PersistentFlags().String(kubernetes.FlagInitToken_worker, "", "Set init-token-worker (Default to new token)")
	setupCmd.PersistentFlags().String(kubernetes.FlagInitToken_master, "", "Set init-token-master (Default to new token)")
	setupCmd.PersistentFlags().String(kubernetes.FlagInitToken_all, "", "Set init-token-all    (Default to new token)")

	RootCmd.AddCommand(setupCmd)
}
