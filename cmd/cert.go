package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jetstack/vault-helper/pkg/cert"
)

// initCmd represents the init command
var CertCmd = &cobra.Command{
	Use: "cert [cert role] [common name] [destination path]",
	// TODO: Make short better
	Short: "Create local key to generate a CSR. Call vault with CSR for specified cert role.",
	Run: func(cmd *cobra.Command, args []string) {
		log, err := LogLevel(cmd)
		if err != nil {
			Must(err)
		}

		i, err := newInstanceToken(cmd)
		if err != nil {
			Must(err)
		}

		if err := i.TokenRenewRun(); err != nil {
			Must(err)
		}

		c := cert.New(log, i)
		if len(args) != 3 {
			Must(fmt.Errorf("wrong number of arguments given. Usage: vault-helper cert [cert role] [common name] [destination path]"))
		}
		if err := setFlagsCert(c, cmd, args); err != nil {
			Must(err)
		}

		if err := c.RunCert(); err != nil {
			Must(err)
		}
	},
}

func init() {
	InitCertCmdFlags(CertCmd)
}

func InitCertCmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Int(cert.FlagKeyBitSize, 2048, "Bit size used for generating key. [int]")
	cmd.Flag(cert.FlagKeyBitSize).Shorthand = "b"

	cmd.PersistentFlags().String(cert.FlagKeyType, "RSA", "Type of key to generate. [string]")
	cmd.Flag(cert.FlagKeyType).Shorthand = "t"

	cmd.PersistentFlags().StringSlice(cert.FlagIpSans, []string{}, "IP sans. [[]string] (default none)")
	cmd.Flag(cert.FlagIpSans).Shorthand = "i"

	cmd.PersistentFlags().StringSlice(cert.FlagSanHosts, []string{}, "Host Sans. [[]string] (default none)")
	cmd.Flag(cert.FlagSanHosts).Shorthand = "s"

	cmd.PersistentFlags().StringSlice(cert.FlagOrganisation, []string{}, "Organisation(s) - i.e. kubernetes groups. [[]string] (default none)")

	cmd.Flag(cert.FlagOrganisation).Shorthand = "n"

	cmd.PersistentFlags().String(cert.FlagOwner, "", "Owner of created file/directories. Uid value also accepted. [string] (default <current user>)")
	cmd.Flag(cert.FlagOwner).Shorthand = "o"

	cmd.PersistentFlags().String(cert.FlagGroup, "", "Group of created file/directories. Gid value also accepted. [string] (default <current user-group)")
	cmd.Flag(cert.FlagGroup).Shorthand = "g"

	instanceTokenFlags(cmd)

	RootCmd.AddCommand(cmd)
}

func setFlagsCert(c *cert.Cert, cmd *cobra.Command, args []string) error {
	vInt, err := cmd.PersistentFlags().GetInt(cert.FlagKeyBitSize)
	if err != nil {
		return fmt.Errorf("error parsing %s [int] '%d': %v", cert.FlagKeyBitSize, vInt, err)
	}
	c.SetBitSize(vInt)

	vStr, err := cmd.PersistentFlags().GetString(cert.FlagKeyType)
	if err != nil {
		return fmt.Errorf("error parsing %s [string] '%s': %v", cert.FlagKeyType, vStr, err)
	}
	c.SetKeyType(vStr)

	vStr, err = cmd.PersistentFlags().GetString(cert.FlagOwner)
	if err != nil {
		return fmt.Errorf("error parsing %s [string] '%s': %v", cert.FlagOwner, vStr, err)
	}
	c.SetOwner(vStr)

	vStr, err = cmd.PersistentFlags().GetString(cert.FlagGroup)
	if err != nil {
		return fmt.Errorf("error parsing %s [string] '%s': %v", cert.FlagGroup, vStr, err)
	}
	c.SetGroup(vStr)

	vSli, err := cmd.PersistentFlags().GetStringSlice(cert.FlagIpSans)
	if err != nil {
		return fmt.Errorf("error parsing %s [[]string] '%s': %v", cert.FlagIpSans, vSli, err)
	}
	c.SetIPSans(vSli)

	vSli, err = cmd.PersistentFlags().GetStringSlice(cert.FlagSanHosts)
	if err != nil {
		return fmt.Errorf("error parsing %s [[]string] '%s': %v", cert.FlagSanHosts, vSli, err)
	}
	c.SetSanHosts(vSli)

	vSli, err = cmd.PersistentFlags().GetStringSlice(cert.FlagOrganisation)
	if err != nil {
		return fmt.Errorf("error parsing %s [[]string] '%s' : %v", cert.FlagOrganisation, vSli, err)
	}
	c.SetOrganisation(vSli)

	abs, err := filepath.Abs(args[2])
	if err != nil {
		return fmt.Errorf("failed to generate absoute path from destination '%s': %v", args[2], err)
	}
	c.SetDestination(abs)

	c.SetRole(args[0])
	c.SetCommonName(args[1])

	return nil
}
