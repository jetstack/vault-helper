package cmd

import (
	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/jetstack-experimental/vault-helper/pkg/read"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var readCmd = &cobra.Command{
	Use:   "read [cluster ID] [vault path] [output file]",
	Short: "Read arbitrary vault path. If no output file specified, output to console.",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()
		logger.Level = logrus.DebugLevel
		log := logrus.NewEntry(logger)

		v, err := vault.NewClient(nil)
		if err != nil {
			logger.Fatal(err)
		}

		r := read.New(v, log)

		if err := r.Run(cmd, args); err != nil {
			logrus.Fatal(err)
		}

	},
}

func init() {
	RootCmd.AddCommand(readCmd)
}
