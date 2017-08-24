package cmd

import (
	"time"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"
)

// initCmd represents the init command
var setupCmd = &cobra.Command{
	Use:   "setup [cluster ID]",
	Short: "Setup kubernetes on a running vault server.",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()

		i, err := RootCmd.PersistentFlags().GetInt("log-level")
		if err != nil {
			logrus.Fatalf("failed to get log level of flag: %s", err)
		}
		if i < 0 || i > 2 {
			logrus.Fatalf("not a valid log level")
		}
		switch i {
		case 0:
			logger.Level = logrus.FatalLevel
		case 1:
			logger.Level = logrus.InfoLevel
		case 2:
			logger.Level = logrus.DebugLevel
		}

		v, err := vault.NewClient(nil)
		if err != nil {
			logger.Fatal(err)
		}

		k := kubernetes.New(v)
		if err != nil {
			logger.Fatal(err)
		}

		if err := k.Run(cmd, args); err != nil {
			logger.Fatal(err)
		}

		for n, t := range k.InitTokens() {
			logrus.Infof(n + "-init_token := " + t)
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
