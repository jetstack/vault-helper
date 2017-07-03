package cmd

import (
	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/kubernetes_pki"
)

// initCmd represents the init command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup kubernetes on a running vault server",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: this should be a cli parameter
		prefix := "kubernetes"
		logrus.Infof("setting up vault on prefix %s", prefix)

		vaultClient, err := vault.NewClient(nil)
		if err != nil {
			logrus.Fatalf("unable to create vault client: ", err)
		}

		kPKI := kubernetes_pki.New(prefix, vaultClient)

		// TODO read env vars and populate
		// kPKI.MaxValidityAdmin ==

		// TODO ensure that it is setup in that way
		// kPKI.Ensure()

		logrus.Debugf("kpki: %#+v", kPKI)

	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
