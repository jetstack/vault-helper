package cmd

import (
	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	"github.com/jetstack-experimental/vault-helper/pkg/read"
)

// initCmd represents the init command
var readCmd = &cobra.Command{
	Use:   "read [vault path]",
	Short: "Read arbitrary vault path. If no output file specified, output to console.",
	Run: func(cmd *cobra.Command, args []string) {
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

		log := logrus.NewEntry(logger)

		v, err := vault.NewClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		r := read.New(v, log)

		if err := r.Run(cmd, args); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	readCmd.PersistentFlags().String(read.FlagOutputPath, "", "Set destination file path of read responce. Output to console if no filepath given (default <console>)")
	readCmd.Flag(read.FlagOutputPath).Shorthand = "d"
	readCmd.PersistentFlags().String(read.FlagField, "", "If included, the raw value of the specified field will be output. If not, output entire responce in JSON (default <all>)")
	readCmd.Flag(read.FlagField).Shorthand = "f"
	readCmd.PersistentFlags().String(read.FlagOwner, "", "Set owner of output file. Uid value also accepted. (default <current user>)")
	readCmd.Flag(read.FlagOwner).Shorthand = "o"
	readCmd.PersistentFlags().String(read.FlagGroup, "", "Set group of output file. Gid value also accepted. (default <current user-group>)")
	readCmd.Flag(read.FlagGroup).Shorthand = "g"

	RootCmd.AddCommand(readCmd)
}
