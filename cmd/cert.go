package cmd

import (
	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/jetstack-experimental/vault-helper/pkg/cert"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var certCmd = &cobra.Command{
	Use: "cert [cert role] [common name] [destination path]",
	// TODO: Make short better
	Short: "Create local key to generate a CSR. Call vault with CSR for specified cert role",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()
		logger.Level = logrus.DebugLevel
		log := logrus.NewEntry(logger)

		v, err := vault.NewClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		c := cert.New(v, log)

		if err := c.Run(cmd, args); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	certCmd.PersistentFlags().Int(cert.FlagKeyBitSize, 2048, "Bit size used for generating key. [int]")
	certCmd.Flag(cert.FlagKeyBitSize).Shorthand = "b"

	certCmd.PersistentFlags().String(cert.FlagKeyType, "RSA", "Type of key to generate. [string]")
	certCmd.Flag(cert.FlagKeyType).Shorthand = "t"

	certCmd.PersistentFlags().StringSlice(cert.FlagIpSans, []string{}, "IP sans. [[]string] (default none)")
	certCmd.Flag(cert.FlagIpSans).Shorthand = "i"

	certCmd.PersistentFlags().StringSlice(cert.FlagSanHosts, []string{}, "Host Sans. [[]string] (default none)")
	certCmd.Flag(cert.FlagSanHosts).Shorthand = "s"

	certCmd.PersistentFlags().String(cert.FlagOwner, "", "Owner of created file/directories. Uid value also accepted. [string] (default <current user>)")
	certCmd.Flag(cert.FlagOwner).Shorthand = "o"

	certCmd.PersistentFlags().String(cert.FlagGroup, "", "Group of created file/directories. Gid value also accepted. [string] (default <current user-group)")
	certCmd.Flag(cert.FlagGroup).Shorthand = "g"

	RootCmd.AddCommand(certCmd)
}
