package instanceToken

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

const FlagTokenRole = "role"

func (i *InstanceToken) Run(cmd *cobra.Command, args []string) error {

	if len(args) > 0 {
		i.clusterID = args[0]
	} else {
		return errors.New("No cluster id was given")
	}

	value, err := cmd.PersistentFlags().GetString(FlagTokenRole)
	if err != nil {
		return fmt.Errorf("Error parsing %s '%s': %s", FlagTokenRole, value, err)
	}

	if value == "" {
		return fmt.Errorf("No token role was given. Token role is required for this command:\n --%s", FlagTokenRole)
	}

	return nil

}
