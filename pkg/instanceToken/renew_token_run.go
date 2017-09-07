package instanceToken

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

const FlagInitRole = "init-role"
const FlagConfigPath = "config-path"

func (i *InstanceToken) Run(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return errors.New("invalid arguments")
	}

	value, err := cmd.PersistentFlags().GetString(FlagInitRole)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %v", FlagInitRole, value, err)
	}
	if value == "" {
		return fmt.Errorf("no token role was given. token role is required for this command: --%s", FlagInitRole)
	}
	i.SetInitRole(value)

	value, err = cmd.Root().Flags().GetString(FlagConfigPath)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %v", FlagConfigPath, value, err)
	}
	if value != "" {
		abs, err := filepath.Abs(value)
		if err != nil {
			return fmt.Errorf("error generating absoute path from path '%s': %v", value, err)
		}
		i.SetVaultConfigPath(abs)
	}

	return i.TokenRenewRun()
}
