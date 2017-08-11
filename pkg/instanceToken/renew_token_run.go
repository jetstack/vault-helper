package instanceToken

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

const FlagTokenRole = "role"
const FlagVaultConfigPath = "config-path"

func (i *InstanceToken) Run(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		i.SetClusterID(args[0])
	} else {
		return errors.New("no cluster id was given")
	}

	value, err := cmd.PersistentFlags().GetString(FlagTokenRole)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %v", FlagTokenRole, value, err)
	}
	if value == "" {
		return fmt.Errorf("nno token role was given. token role is required for this command: --%s", FlagTokenRole)
	}
	i.SetRole(value)

	value, err = cmd.Root().Flags().GetString(FlagVaultConfigPath)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %v", FlagVaultConfigPath, value, err)
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
