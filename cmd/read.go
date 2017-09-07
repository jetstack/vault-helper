package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack-experimental/vault-helper/pkg/read"
)

// initCmd represents the init command
var readCmd = &cobra.Command{
	Use:   "read [vault path]",
	Short: "Read arbitrary vault path. If no output file specified, output to console.",
	Run: func(cmd *cobra.Command, args []string) {
		log := LogLevel(cmd)

		i, err := newInstanceToken(cmd)
		if err != nil {
			i.Log.Fatal(err)
		}

		if err := i.Run(cmd, args); err != nil {
			i.Log.Fatal(err)
		}

		r := read.New(log, i)

		if err := r.Run(cmd, args); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	instanceTokenFlags(readCmd)

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
