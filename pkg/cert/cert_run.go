package cert

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

const FlagKeyBitSize = "key-bit-size"
const FlagKeyType = "key-type"
const FlagIpSans = "ip-sans"
const FlagSanHosts = "san-hosts"
const FlagOwner = "owner"
const FlagGroup = "group"

func (c *Cert) Run(cmd *cobra.Command, args []string) error {

	if len(args) != 3 {
		return errors.New("Wrong number of arguments given.\n    Usage: vault-helper cert [cert role] [common name] [destination path]")
	}

	c.SetRole(args[0])
	c.SetCommonName(args[1])
	c.SetDestination(args[2])

	vInt, err := cmd.PersistentFlags().GetInt(FlagKeyBitSize)
	if err != nil {
		return fmt.Errorf("Error parsing %s [int] '%s': %s", FlagKeyBitSize, vInt, err)
	}
	c.SetBitSize(vInt)

	vStr, err := cmd.PersistentFlags().GetString(FlagKeyType)
	if err != nil {
		return fmt.Errorf("Error parsing %s [string] '%s': %s", FlagKeyType, vStr, err)
	}
	c.SetKeyType(vStr)

	vStr, err = cmd.PersistentFlags().GetString(FlagOwner)
	if err != nil {
		return fmt.Errorf("Error parsing %s [string] '%s': %s", FlagOwner, vStr, err)
	}
	c.SetOwner(vStr)

	vStr, err = cmd.PersistentFlags().GetString(FlagGroup)
	if err != nil {
		return fmt.Errorf("Error parsing %s [string] '%s': %s", FlagGroup, vStr, err)
	}
	c.SetGroup(vStr)

	vSli, err := cmd.PersistentFlags().GetStringSlice(FlagIpSans)
	if err != nil {
		return fmt.Errorf("Error parsing %s [[]string] '%s': %s", FlagIpSans, vSli, err)
	}
	c.SetIPSans(vSli)

	vSli, err = cmd.PersistentFlags().GetStringSlice(FlagSanHosts)
	if err != nil {
		return fmt.Errorf("Error parsing %s [[]string] '%s': %s", FlagSanHosts, vSli, err)
	}
	c.SetSanHosts(vSli)

	return c.RunCert()

}
