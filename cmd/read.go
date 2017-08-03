package cmd

import (
	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/jetstack-experimental/vault-helper/pkg/read"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var readCmd = &cobra.Command{
	Use:   "read [vault path]",
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
	readCmd.PersistentFlags().String(read.FlagOutputPath, "", "Set destination file path of read responce. Output to console if no filepath given (default <console>)")
	readCmd.Flag(read.FlagOutputPath).Shorthand = "d"
	readCmd.PersistentFlags().String(read.FlagField, "", "If included, the raw value of the specified field will be output. If not, output entire responce in JSON (default <all>)")
	readCmd.Flag(read.FlagField).Shorthand = "f"
	readCmd.PersistentFlags().String(read.FlagOwner, "", "Set owner of output file. (default <current user>)")
	readCmd.Flag(read.FlagOwner).Shorthand = "o"
	readCmd.PersistentFlags().String(read.FlagGroup, "", "Set group of output file. (default <current user-group>)")
	readCmd.Flag(read.FlagGroup).Shorthand = "g"

	RootCmd.AddCommand(readCmd)
}
