package kubernetes

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

func NewPKI(k *Kubernetes, pkiName string) *PKI {
	return &PKI{
		pkiName:         pkiName,
		kubernetes:      k,
		MaxLeaseTTL:     k.MaxValidityCA,
		DefaultLeaseTTL: k.MaxValidityCA,
	}
}

type PKI struct {
	pkiName    string
	kubernetes *Kubernetes

	MaxLeaseTTL     time.Duration
	DefaultLeaseTTL time.Duration
}

func (p *PKI) TuneMount(mount *vault.MountOutput) error {

	tuneMountRequired := false

	if mount.Config.DefaultLeaseTTL != int(p.DefaultLeaseTTL.Seconds()) {
		tuneMountRequired = true
	}
	if mount.Config.MaxLeaseTTL != int(p.MaxLeaseTTL.Seconds()) {
		tuneMountRequired = true
	}

	if tuneMountRequired {
		mountConfig := p.getMountConfigInput()
		err := p.kubernetes.vaultClient.Sys().TuneMount(p.Path(), mountConfig)
		if err != nil {
			return fmt.Errorf("error tuning mount config: %s", err.Error())
		}
		logrus.Infof("Tuned Mount: %s", p.pkiName)
		return nil
	}
	logrus.Infof("No tune required: %s", p.pkiName)

	return nil

}

func (p *PKI) Ensure() error {

	mount, err := GetMountByPath(p.kubernetes.vaultClient, p.Path())
	if err != nil {
		return err
	}

	if mount == nil {
		logrus.Infof("No mounts found for: %s", p.pkiName)
		err := p.kubernetes.vaultClient.Sys().Mount(
			p.Path(),
			&vault.MountInput{
				Description: "Kubernetes " + p.kubernetes.clusterID + "/" + p.pkiName + " CA",
				Type:        "pki",
				Config:      p.getMountConfigInput(),
			},
		)
		if err != nil {
			return fmt.Errorf("error creating mount: %s", err)
		}
		logrus.Infof("Mounted: %s", p.pkiName)

		mount, err = GetMountByPath(p.kubernetes.vaultClient, p.Path())
		if err != nil {
			return err
		}

	} else {
		if mount.Type != "pki" {
			return fmt.Errorf("Mount '%s' already existing with wrong type '%s'", p.Path(), mount.Type)
		}
		logrus.Infof("Mount '%s' already existing", p.Path())
	}

	if mount != nil {
		err = p.TuneMount(mount)
		if err != nil {
			logrus.Fatalf("Tuning Error")
			return err
		}
	}

	return nil
}

func (p *PKI) WriteRole(role *pkiRole) error {

	path := filepath.Join(p.Path(), "roles", role.Name)

	_, err := p.kubernetes.vaultClient.Logical().Write(path, role.Data)
	if err != nil {
		return fmt.Errorf("error writting role '%s' to '%s': %s", role.Name, p.Path(), err)
	}

	return nil
}

func (p *PKI) Path() string {
	return filepath.Join(p.kubernetes.Path(), "pki", p.pkiName)
}

func (p *PKI) getMountConfigInput() vault.MountConfigInput {
	return vault.MountConfigInput{
		DefaultLeaseTTL: p.getDefaultLeaseTTL(),
		MaxLeaseTTL:     p.getMaxLeaseTTL(),
	}
}

func (p *PKI) getDefaultLeaseTTL() string {
	return fmt.Sprintf("%d", int(p.DefaultLeaseTTL.Seconds()))
}

func (p *PKI) getMaxLeaseTTL() string {
	return fmt.Sprintf("%d", int(p.MaxLeaseTTL.Seconds()))
}

func (p *PKI) getTokenPolicyExists(name string) (bool, error) {

	policy, err := p.kubernetes.vaultClient.Sys().GetPolicy(name)
	if err != nil {
		return false, err
	}

	if policy == "" {
		logrus.Infof("Policy Not Found: %s", name)
		return false, nil
	}

	logrus.Infof("Policy Found: %s", name)

	return true, nil
}
