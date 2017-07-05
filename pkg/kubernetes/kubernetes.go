package kubernetes

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"
	vault "github.com/hashicorp/vault/api"
)

type Backend interface {
	Ensure() error
	Path() string
}

type Kubernetes struct {
	clusterID string // clusterID is required parameter, lowercase only, [a-z0-9-]+

	etcdKubernetesPKI *PKI
	etcdOverlayPKI    *PKI
	kubernetesPKI     *PKI
	secretsGeneric    *Generic
}

var _ Backend = &PKI{}
var _ Backend = &Generic{}

func IsValidClusterID(clusterID string) error {
	if !unicode.IsLetter([]rune(clusterID)[0]) {
		return errors.New("First character is not a valid character")
	}

	f := func(r rune) bool {
		return ((r < 'a' || r > 'z') && (r < '0' || r > '9')) && r != '-'
	}

	logrus.Infof(clusterID)
	if strings.IndexFunc(clusterID, f) != -1 {
		logrus.Infof("Invalid cluster ID")
		return errors.New("Not a valid cluster ID name")
	}

	logrus.Infof("Valid string")

	return nil

}

func New(clusterID string) (*Kubernetes, error) {

	err := IsValidClusterID(clusterID)
	if err != nil {
		return nil, errors.New("Not a valid cluster ID")
	}

	k := &Kubernetes{
		clusterID: clusterID,
	}

	vaultClient, err := vault.NewClient(nil)
	if err != nil {
		return nil, errors.New("Unable to create vault client")
	}

	k.etcdKubernetesPKI = NewPKI(k, "etcd-k8s", vaultClient)
	k.etcdOverlayPKI = NewPKI(k, "etcd-overlay", vaultClient)
	k.kubernetesPKI = NewPKI(k, "k8s", vaultClient)

	k.secretsGeneric = NewGeneric(k, vaultClient)

	return k, nil
}

func (k *Kubernetes) backends() []Backend {
	return []Backend{
		k.etcdKubernetesPKI,
		k.etcdOverlayPKI,
		k.kubernetesPKI,
	}
}

func (k *Kubernetes) Ensure() error {
	var result error
	for _, backend := range k.backends() {
		if err := backend.Ensure(); err != nil {
			result = multierror.Append(result, fmt.Errorf("backend %s: %s", backend.Path(), err))
		}
	}
	return result
}

func (k *Kubernetes) Path() string {
	return k.clusterID
}

func NewGeneric(k *Kubernetes, vaultClient *vault.Client) *Generic {
	return &Generic{
		kubernetes:  k,
		vaultClient: vaultClient,
	}
}

func NewPKI(k *Kubernetes, pkiName string, vaultClient *vault.Client) *PKI {
	return &PKI{
		pkiName:         pkiName,
		kubernetes:      k,
		vaultClient:     vaultClient,
		MaxLeaseTTL:     time.Hour * 24 * 60,
		DefaultLeaseTTL: time.Hour * 24 * 30,
	}
}

type PKI struct {
	pkiName     string
	kubernetes  *Kubernetes
	vaultClient *vault.Client

	MaxLeaseTTL     time.Duration
	DefaultLeaseTTL time.Duration
}

func TuneMount(p *PKI, mount *vault.MountOutput) error {

	tuneMountRequired := false

	if mount.Config.DefaultLeaseTTL != int(p.DefaultLeaseTTL.Seconds()) {
		tuneMountRequired = true
	}
	if mount.Config.MaxLeaseTTL != int(p.MaxLeaseTTL.Seconds()) {
		tuneMountRequired = true
	}

	if tuneMountRequired {
		mountConfig := p.getMountConfigInput()
		err := p.vaultClient.Sys().TuneMount(p.Path(), mountConfig)
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

	mount, err := GetMountByPath(p.vaultClient, p.Path())
	if err != nil {
		return err
	}

	if mount == nil {
		logrus.Infof("No mounts found for: %s", p.pkiName)
		err := p.vaultClient.Sys().Mount(
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

		mount, err = GetMountByPath(p.vaultClient, p.Path())
		if err != nil {
			return err
		}

	} else {
		if mount.Type != "pki" {
			return fmt.Errorf("Mount '%s' already existing with wrong type '%s'", p.Path(), mount.Type)
		}
		return fmt.Errorf("Mount '%s' already existing", p.Path())
	}

	err = TuneMount(p, mount)
	if err != nil {
		logrus.Fatalf("Tuning Error")
		return err
	}

	return nil
}

func (p *PKI) Path() string {
	return filepath.Join(p.kubernetes.Path(), "pki", p.pkiName)
}

type Generic struct {
	kubernetes  *Kubernetes
	vaultClient *vault.Client
}

func (g *Generic) Ensure() error {
	return errors.New("implement me")
}

func (g *Generic) Path() string {
	return filepath.Join(g.kubernetes.Path(), "generic")
}

func GetMountByPath(vaultClient *vault.Client, mountPath string) (*vault.MountOutput, error) {

	mounts, err := vaultClient.Sys().ListMounts()
	if err != nil {
		return nil, fmt.Errorf("error listing mounts: %s", err)
	}

	var mount *vault.MountOutput
	for key, _ := range mounts {
		if filepath.Clean(key) == filepath.Clean(mountPath) {
			mount = mounts[key]
			break
		}
	}

	return mount, nil
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

func WriteRoles(p *PKI, writeData map[string]interface{}, role string) error {

	rolePath := filepath.Join(p.Path(), "roles", role)
	_, err := p.vaultClient.Logical().Write(rolePath, writeData)

	if err != nil {
		return fmt.Errorf("error writting role data: %s", err)
	}

	return nil
}

func WriteTokenRoles(k *Kubernetes, p *PKI, writeData map[string]interface{}, role string) error {

	rolePath := filepath.Join("auth/token/roles", k.clusterID+"-"+role)
	_, err := p.vaultClient.Logical().Write(rolePath, writeData)

	if err != nil {
		return fmt.Errorf("error writting role data: %s", err)
	}

	return nil
}

func WritePolicy(p *PKI, policy_name, policy string) error {

	err := p.vaultClient.Sys().PutPolicy(policy_name, policy)
	if err != nil {
		return fmt.Errorf("error writting policy: %s", err)
	}

	logrus.Infof("Policy written")

	return nil
}

func getTokenPolicyExists(p *PKI, name string) (bool, error) {

	policy, err := p.vaultClient.Sys().GetPolicy(name)
	if err != nil {
		return false, err
	}

	if policy == "" {
		return false, nil
	}

	return true, nil
}

func (k *Kubernetes) GenerateSecretsMount() error {

	secrets_path := filepath.Join(k.clusterID, "secrets")

	mount, err := GetMountByPath(k.secretsGeneric.vaultClient, secrets_path)
	if err != nil {
		return err
	}

	if mount == nil {
		logrus.Infof("No secrects mount found for: %s", secrets_path)
		err = k.secretsGeneric.vaultClient.Sys().Mount(
			secrets_path,
			&vault.MountInput{
				Description: "Kubernetes " + k.clusterID + " secrets",
				Type:        "generic",
			},
		)

		if err != nil {
			return fmt.Errorf("error creating mount: %s", err)
		}

		logrus.Infof("Mounted secrets")

		reader := rand.Reader
		bitSize := 4096
		key, err := rsa.GenerateKey(reader, bitSize)

		if err != nil {
			return fmt.Errorf("error generating rsa key: %s", err)
		}

		writeData := map[string]interface{}{
			"key": key,
		}

		secrets_path = filepath.Join(secrets_path, "service-accounts")

		_, err = k.secretsGeneric.vaultClient.Logical().Write(secrets_path, writeData)

		if err != nil {
			logrus.Fatal("Error writting key to secrets", err)
		}
		logrus.Infof("Key written to secrets")

	}

	return nil
}
