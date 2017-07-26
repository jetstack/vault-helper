package kubernetes

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

const FlagMaxValidityAdmin = "max-validity-admin"
const FlagMaxValidityCA = "max-validity-ca"
const FlagMaxValidityComponents = "max-validity-components"

const FlagInitToken_etcd = "init-token-etcd"
const FlagInitToken_all = "init-token-all"
const FlagInitToken_master = "init-token-master"
const FlagInitToken_worker = "init-token-worker"

func (k *Kubernetes) Run(cmd *cobra.Command, args []string) error {

	if value, err := cmd.PersistentFlags().GetDuration(FlagMaxValidityComponents); err != nil {
		if err != nil {
			return fmt.Errorf("error parsing %s '%s': %s", FlagMaxValidityComponents, value, err)
		}
		k.MaxValidityComponents = value
	}

	if value, err := cmd.PersistentFlags().GetDuration(FlagMaxValidityAdmin); err != nil {
		if err != nil {
			return fmt.Errorf("error parsing %s '%s': %s", FlagMaxValidityAdmin, value, err)
		}
		k.MaxValidityAdmin = value
	}

	if value, err := cmd.PersistentFlags().GetDuration(FlagMaxValidityCA); err != nil {
		if err != nil {
			return fmt.Errorf("error parsing %s '%s': %s", FlagMaxValidityCA, value, err)
		}
		k.MaxValidityCA = value
	}

	// Init token flags
	value, err := cmd.PersistentFlags().GetString(FlagInitToken_etcd)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", FlagInitToken_etcd, value, err)
	}
	k.FlagInitTokens.etcd = value

	value, err = cmd.PersistentFlags().GetString(FlagInitToken_master)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", FlagInitToken_master, value, err)
	}
	k.FlagInitTokens.master = value

	value, err = cmd.PersistentFlags().GetString(FlagInitToken_worker)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", FlagInitToken_worker, value, err)
	}
	k.FlagInitTokens.worker = value

	value, err = cmd.PersistentFlags().GetString(FlagInitToken_all)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", FlagInitToken_all, value, err)
	}
	k.FlagInitTokens.all = value

	// TODO: ensure CA >> COMPONENTS/ADMIN

	if len(args) > 0 {
		k.clusterID = args[0]
	} else {
		return errors.New("no cluster id was given")
	}

	return k.Ensure()
}
