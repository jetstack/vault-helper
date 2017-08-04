package read

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

const FlagOutputPath = "dest-path"
const FlagField = "field"
const FlagOwner = "owner"
const FlagGroup = "group"

func (r *Read) Run(cmd *cobra.Command, args []string) error {

	if len(args) > 1 {
		return errors.New("incorrect number of arguments given. usage: vault-helper read [vault path]")
	}

	r.SetVaultPath(args[0])

	value, err := cmd.PersistentFlags().GetString(FlagOutputPath)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", FlagOutputPath, value, err)
	}
	if value != "" {
		abs, err := filepath.Abs(value)
		if err != nil {
			return fmt.Errorf("error generating absoute path from destination '%s': %s", value, err)
		}
		r.SetFilePath(abs)
	}

	value, err = cmd.PersistentFlags().GetString(FlagField)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", FlagField, value, err)
	}
	if value != "" {
		r.SetFieldName(value)
	}

	value, err = cmd.PersistentFlags().GetString(FlagOwner)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", FlagOwner, value, err)
	}
	if value != "" {
		r.SetOwner(value)
	}

	value, err = cmd.PersistentFlags().GetString(FlagGroup)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", FlagGroup, value, err)
	}
	if value != "" {
		r.SetGroup(value)
	}

	return r.RunRead()
}
