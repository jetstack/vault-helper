package kubeconfig

import (
	//"errors"
	"fmt"
	"path/filepath"

	"github.com/jetstack-experimental/vault-helper/pkg/cert"
	"github.com/spf13/cobra"
)

func (u *Kubeconfig) Run(cmd *cobra.Command, args []string) error {

	abs, err := filepath.Abs(args[3])
	if err != nil {
		return fmt.Errorf("Error generating absoute path from destination '%s':\n%s", args[3], err)
	}
	u.SetFilePath(abs)

	args = args[:3]

	c := cert.New(u.vaultClient, u.Log)
	u.SetCert(c)

	if err := c.Run(cmd, args); err != nil {
		u.Log.Fatal(err)
	}

	return u.RunKube()
}
