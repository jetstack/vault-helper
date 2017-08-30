package cmd

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
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

	RootCmd.PersistentFlags().Int("log-level", 1, "Set the log level of output. 0-Fatal 1-Info 2-Debug")
	RootCmd.Flag("log-level").Shorthand = "l"

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func LogLevel(cmd *cobra.Command) *logrus.Entry {
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

	return logrus.NewEntry(logger)
}
