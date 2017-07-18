package kubernetes

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"
	vault "github.com/hashicorp/vault/api"
)

type Backend interface {
	Ensure() error
	Path() string
}

type VaultLogical interface {
	Write(path string, data map[string]interface{}) (*vault.Secret, error)
	Read(path string) (*vault.Secret, error)
}

type VaultSys interface {
	ListMounts() (map[string]*vault.MountOutput, error)
	ListPolicies() ([]string, error)

	Mount(path string, mountInfo *vault.MountInput) error
	PutPolicy(name, rules string) error
	TuneMount(path string, config vault.MountConfigInput) error
	GetPolicy(name string) (string, error)
}

type VaultAuth interface {
	Token() VaultToken
}

type VaultToken interface {
	CreateOrphan(opts *vault.TokenCreateRequest) (*vault.Secret, error)
}

type Vault interface {
	Logical() VaultLogical
	Sys() VaultSys
	Auth() VaultAuth
}

type realVault struct {
	c *vault.Client
}

type realVaultAuth struct {
	a *vault.Auth
}

func (rv *realVault) Auth() VaultAuth {
	return &realVaultAuth{a: rv.c.Auth()}
}
func (rv *realVault) Sys() VaultSys {
	return rv.c.Sys()
}
func (rv *realVault) Logical() VaultLogical {
	return rv.c.Logical()
}

func (rva *realVaultAuth) Token() VaultToken {
	return rva.a.Token()
}

func RealVaultFromAPI(vaultClient *vault.Client) Vault {
	return &realVault{c: vaultClient}
}

type Kubernetes struct {
	clusterID   string // clusterID is required parameter, lowercase only, [a-z0-9-]+
	vaultClient Vault

	etcdKubernetesPKI *PKI
	etcdOverlayPKI    *PKI
	kubernetesPKI     *PKI
	secretsGeneric    *Generic
}

type Policy struct {
	policy_name string
	rules       string
	role        string
	kubernetes  *Kubernetes
}

type TokenRole struct {
	role_name  string
	writeData  map[string]interface{}
	kubernetes *Kubernetes
}

type ComponentRole struct {
	component  string
	role       string
	writeData  map[string]interface{}
	kubernetes *Kubernetes
}

type InitTokenPolicy struct {
	policy_name string
	role_name   string
	initToken   string
	kubernetes  *Kubernetes
}

func (t *TokenRole) WriteTokenRole() error {
	rolePath := filepath.Join("auth/token/roles", t.kubernetes.clusterID+"-"+t.role_name)
	rolePath = filepath.Clean(rolePath)
	_, err := t.kubernetes.vaultClient.Logical().Write(rolePath, t.writeData)

	if err != nil {
		return fmt.Errorf("error writting role data: %s", err)
	}

	return nil
}

func (k *Kubernetes) NewTokenRole(role_name string, writeData map[string]interface{}) *TokenRole {
	return &TokenRole{
		role_name:  role_name,
		writeData:  writeData,
		kubernetes: k,
	}

}

func (k *Kubernetes) NewComponentRole(component, role string, writeData map[string]interface{}) *ComponentRole {
	return &ComponentRole{
		component:  component,
		role:       role,
		writeData:  writeData,
		kubernetes: k,
	}

}

func (i *ComponentRole) WriteComponentRole() error {
	path := filepath.Join(i.kubernetes.clusterID, "pki", i.component, "roles", i.role)

	_, err := i.kubernetes.vaultClient.Logical().Write(path, i.writeData)

	if err != nil {
		return fmt.Errorf("error writting role data: %s", err)
	}

	return nil
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

func New(vaultClient Vault, clusterID string) (*Kubernetes, error) {

	err := IsValidClusterID(clusterID)
	if err != nil {
		return nil, errors.New("Not a valid cluster ID")
	}

	k := &Kubernetes{
		clusterID:   clusterID,
		vaultClient: vaultClient,
	}

	k.etcdKubernetesPKI = NewPKI(k, "etcd-k8s")
	k.etcdOverlayPKI = NewPKI(k, "etcd-overlay")
	k.kubernetesPKI = NewPKI(k, "k8s")

	k.secretsGeneric = k.NewGeneric()

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

func (k *Kubernetes) NewGeneric() *Generic {
	return &Generic{
		kubernetes: k,
		initTokens: make(map[string]string),
	}
}

func GetMountByPath(vaultClient Vault, mountPath string) (*vault.MountOutput, error) {

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

func (k *Kubernetes) NewPolicy(policy_name, rules, role string) *Policy {
	return &Policy{
		policy_name: policy_name,
		rules:       rules,
		role:        role,
		kubernetes:  k,
	}

}

func (p *Policy) WritePolicy() error {

	err := p.kubernetes.vaultClient.Sys().PutPolicy(p.policy_name, p.rules)
	if err != nil {
		return fmt.Errorf("error writting policy: %s", err)
	}
	logrus.Infof("Policy written: %s", p.policy_name)

	return nil

}

func (p *Policy) CreateTokenCreater() error {
	createrRule := "path \"auth/token/create/" + p.kubernetes.clusterID + "-" + p.role + "+\" {\n    capabilities = [\"create\",\"read\",\"update\"]\n}"
	err := p.kubernetes.vaultClient.Sys().PutPolicy(p.policy_name+"-creator", createrRule)
	if err != nil {
		return fmt.Errorf("error writting creator policy: %s", err)
	}
	logrus.Infof("Creator policy written: %s", p.policy_name)

	return nil
}

func (k *Kubernetes) NewInitToken(policy_name, role_name string) *InitTokenPolicy {
	return &InitTokenPolicy{
		policy_name: policy_name,
		role_name:   role_name,
		initToken:   randomUUID(),
		kubernetes:  k,
	}
}

func (k *Kubernetes) InitTokens() map[string]string {

	return k.secretsGeneric.initTokens
}
