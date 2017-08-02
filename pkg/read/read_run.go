package read

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

func (r *Read) Run(cmd *cobra.Command, args []string) error {

	if len(args) < 2 || len(args) > 3 {
		return errors.New("Incorrect number of arguments given.\n    Usage: vault-helper read [cluster ID] [vault path] [output file]")
	}

	r.SetClusterID(args[0])
	r.SetVaultPath(args[1])

	if len(args) > 2 {
		abs, err := filepath.Abs(args[2])
		if err != nil {
			return fmt.Errorf("Error generating absoute path from output file path '%s':\n%s", args[2], err)
		}
		r.SetFilePath(abs)
	}

	return r.RunRead()
}
