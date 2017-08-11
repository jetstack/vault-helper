package cert

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/jetstack-experimental/vault-helper/pkg/instanceToken"
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
		return errors.New("wrong number of arguments given. Usage: vault-helper cert [cert role] [common name] [destination path]")
	}

	c.SetRole(args[0])
	c.SetCommonName(args[1])

	abs, err := filepath.Abs(args[2])
	if err != nil {
		return fmt.Errorf("failed to generate absoute path from destination '%s': %s", args[2], err)
	}
	c.SetDestination(abs)

	vInt, err := cmd.PersistentFlags().GetInt(FlagKeyBitSize)
	if err != nil {
		return fmt.Errorf("error parsing %s [int] '%d': %s", FlagKeyBitSize, vInt, err)
	}
	c.SetBitSize(vInt)

	vStr, err := cmd.PersistentFlags().GetString(FlagKeyType)
	if err != nil {
		return fmt.Errorf("error parsing %s [string] '%s': %s", FlagKeyType, vStr, err)
	}
	c.SetKeyType(vStr)

	vStr, err = cmd.PersistentFlags().GetString(FlagOwner)
	if err != nil {
		return fmt.Errorf("error parsing %s [string] '%s': %s", FlagOwner, vStr, err)
	}
	c.SetOwner(vStr)

	vStr, err = cmd.PersistentFlags().GetString(FlagGroup)
	if err != nil {
		return fmt.Errorf("error parsing %s [string] '%s': %s", FlagGroup, vStr, err)
	}
	c.SetGroup(vStr)

	vSli, err := cmd.PersistentFlags().GetStringSlice(FlagIpSans)
	if err != nil {
		return fmt.Errorf("error parsing %s [[]string] '%s': %s", FlagIpSans, vSli, err)
	}
	c.SetIPSans(vSli)

	vSli, err = cmd.PersistentFlags().GetStringSlice(FlagSanHosts)
	if err != nil {
		return fmt.Errorf("error parsing %s [[]string] '%s': %s", FlagSanHosts, vSli, err)
	}
	c.SetSanHosts(vSli)

	value, err := cmd.Root().Flags().GetString(instanceToken.FlagVaultConfigPath)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", instanceToken.FlagVaultConfigPath, value, err)
	}
	if value != "" {
		abs, err := filepath.Abs(value)
		if err != nil {
			return fmt.Errorf("error generating absoute path from path '%s': %s", value, err)
		}
		c.SetVaultConfigPath(abs)
	}

	return c.RunCert()

}