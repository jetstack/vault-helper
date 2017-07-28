package cert

import (
	"errors"
	//"fmt"

	"github.com/spf13/cobra"
)

const FlagKeyBitSize = "key-bit-size"
const FlagKeyType = "key-type"
const FlagIpSans = "ip-sans"
const FlagSanHosts = "san-hosts"
const FlagOwner = "owner"
const FlagGroup = "group"

func (i *Cert) Run(cmd *cobra.Command, args []string) error {

	if len(args) != 3 {
		return errors.New("Wrong number of arguments given.\n           Usage vault-helper cert [cert role] [common name] [destination path]")
	}

	return nil

}
