package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jetstack-experimental/vault-helper/pkg/instanceToken"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "vault-helper",
	Short: "Automates PKI tasks using Hashicorp's Vault as a backend.",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	RootCmd.PersistentFlags().String(instanceToken.FlagVaultConfigPath, "/etc/vault", "Set config path to directory with tokens")
	RootCmd.Flag(instanceToken.FlagVaultConfigPath).Shorthand = "p"

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
