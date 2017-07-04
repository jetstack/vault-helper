package cmd

import (
	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	"gitlab.jetstack.net/jetstack-experimental/vault-helper/pkg/kubernetes_pki"
	"time"
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

		vaultClient.Sys().Mount(
			"test3/pki/etcd-k8s/",
			&vault.MountInput{
				Description: "Kubernetes test3/etcd-k8s CA",
				Type:        "pki",
			},
		)

		vaultClient.Sys().Mount(
			"test/pki/etcd-overlay/",
			&vault.MountInput{
				Description: "Kubernetes test3/etcd-overlay CA",
				Type:        "pki",
			},
		)

		vaultClient.Sys().Mount(
			"test3/pki/k8s/",
			&vault.MountInput{
				Description: "Kubernetes test3/k8s CA",
				Type:        "pki",
			},
		)

		kPKI := kubernetes_pki.New(prefix, vaultClient)

		// TODO read env vars and populate
		kPKI.MaxValidityAdmin = time.Hour * 24 * 60

		// TODO ensure that it is setup in that way
		// kPKI.Ensure()
		logrus.Debugf("kpki: %#+v", kPKI)

	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
