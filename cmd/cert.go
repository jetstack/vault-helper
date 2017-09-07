package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack-experimental/vault-helper/pkg/cert"
)

// initCmd represents the init command
var certCmd = &cobra.Command{
	Use: "cert [cert role] [common name] [destination path]",
	// TODO: Make short better
	Short: "Create local key to generate a CSR. Call vault with CSR for specified cert role.",
	Run: func(cmd *cobra.Command, args []string) {
		log := LogLevel(cmd)

		i, err := newInstanceToken(cmd)
		if err != nil {
			i.Log.Fatal(err)
		}

		if err := i.Run(cmd, args); err != nil {
			i.Log.Fatal(err)
		}

		c := cert.New(log, i)

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

	instanceTokenFlags(certCmd)

	RootCmd.AddCommand(certCmd)
}
