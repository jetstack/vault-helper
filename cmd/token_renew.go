// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var renewtokenCmd = &cobra.Command{
	Use:   "renew-token",
	Short: "Renew token on vault server.",
	Run: func(cmd *cobra.Command, args []string) {

		i, err := newInstanceToken(cmd)
		if err != nil {
			Must(err)
		}

		if err := i.TokenRenewRun(); err != nil {
			Must(err)
		}
	},
}

func init() {
	instanceTokenFlags(renewtokenCmd)
	RootCmd.AddCommand(renewtokenCmd)
}
