// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack/vault-helper/pkg/instanceToken"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "vault-helper",
	Short: "Automates PKI tasks using Hashicorp's Vault as a backend.",
}

var Must = func(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().Int("log-level", 1, "Set the log level of output. 0-Fatal 1-Info 2-Debug")
	RootCmd.Flag("log-level").Shorthand = "l"
}

func instanceTokenFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(instanceToken.FlagConfigPath, "p", "/etc/vault", "Set config path to directory with tokens")
	cmd.PersistentFlags().StringP(instanceToken.FlagInitRole, "r", "", "Set role of token to renew. (default *no role*)")
}

func newInstanceToken(cmd *cobra.Command) (*instanceToken.InstanceToken, error) {
	var result *multierror.Error

	log, err := LogLevel(cmd)
	if err != nil {
		return nil, err
	}

	v, err := vault.NewClient(nil)
	if err != nil {
		return nil, err
	}

	i := instanceToken.New(v, log)

	initRole, err := cmd.Flags().GetString(instanceToken.FlagInitRole)
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("error parsing %s '%s': %v", instanceToken.FlagInitRole, initRole, err))
	}
	if initRole == "" {
		//Read env variable
		initRole = os.Getenv("VAULT_INIT_ROLE")
		if initRole == "" {
			result = multierror.Append(result, fmt.Errorf("no token role was given. token role is required for this command: --%s", instanceToken.FlagInitRole))
		}
	}
	i.SetInitRole(initRole)

	vaultConfigPath, err := cmd.Flags().GetString(instanceToken.FlagConfigPath)
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("error parsing %s '%s': %v", instanceToken.FlagConfigPath, vaultConfigPath, err))
	}
	if vaultConfigPath != "" {
		abs, err := filepath.Abs(vaultConfigPath)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error generating absoute path from path '%s': %v", vaultConfigPath, err))
		}
		i.SetVaultConfigPath(abs)
	}

	return i, result.ErrorOrNil()
}

func LogLevel(cmd *cobra.Command) (*logrus.Entry, error) {
	logger := logrus.New()

	i, err := RootCmd.PersistentFlags().GetInt("log-level")
	if err != nil {
		return nil, fmt.Errorf("failed to get log level of flag: %s", err)
	}
	if i < 0 || i > 2 {
		return nil, fmt.Errorf("not a valid log level")
	}
	switch i {
	case 0:
		logger.Level = logrus.FatalLevel
	case 1:
		logger.Level = logrus.InfoLevel
	case 2:
		logger.Level = logrus.DebugLevel
	}

	return logrus.NewEntry(logger), nil
}
