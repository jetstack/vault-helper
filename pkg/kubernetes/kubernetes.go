package kubernetes

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	vault "github.com/hashicorp/vault/api"
)

type Backend interface {
	Ensure() error
	Path() string
}

type Kubernetes struct {
	clusterID   string // clusterID is required parameter, lowercase only, [a-z0-9-]+
	vaultClient *vault.Client

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

type InitTokenPolicy struct {
	policy_name string
	role_name   string
	initToken   string
	kubernetes  *Kubernetes
}

func WriteComponentRole(path string, writeData map[string]interface{}, k *Kubernetes) error {
	_, err := k.vaultClient.Logical().Write(path, writeData)

	if err != nil {
		return fmt.Errorf("error writting role data: %s", err)
	}

	return nil
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

func NewTokenRole(role_name string, writeData map[string]interface{}, k *Kubernetes) *TokenRole {
	return &TokenRole{
		role_name:  role_name,
		writeData:  writeData,
		kubernetes: k,
	}

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

func New(vaultClient *vault.Client, clusterID string) (*Kubernetes, error) {

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

	k.secretsGeneric = NewGeneric(k)

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

func NewGeneric(k *Kubernetes) *Generic {
	return &Generic{
		kubernetes: k,
	}
}

type Generic struct {
	kubernetes *Kubernetes
}

func (g *Generic) Ensure() error {
	err := g.kubernetes.GenerateSecretsMount()
	return err
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

func NewPolicy(policy_name, rules, role string, k *Kubernetes) *Policy {
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

func (k *Kubernetes) GenerateSecretsMount() error {

	secrets_path := filepath.Join(k.clusterID, "secrets")

	mount, err := GetMountByPath(k.vaultClient, secrets_path)
	if err != nil {
		return err
	}

	if mount == nil {
		logrus.Infof("No secrects mount found for: %s", secrets_path)
		err = k.vaultClient.Sys().Mount(
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

		err = writeKey(k, secrets_path)
		if err != nil {
			return fmt.Errorf("error creating mount: %s", err)
		}

	}

	return nil
}

func writeKey(k *Kubernetes, secrets_path string) error {

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

	_, err = k.vaultClient.Logical().Write(secrets_path, writeData)

	if err != nil {
		logrus.Fatal("Error writting key to secrets", err)
	}
	logrus.Infof("Key written to secrets")

	return nil
}

func randomUUID() string {
	uuID := uuid.New()
	return string(uuID[:])
}

func NewInitToken(policy_name, role_name string, k *Kubernetes) *InitTokenPolicy {
	return &InitTokenPolicy{
		policy_name: policy_name,
		role_name:   role_name,
		initToken:   randomUUID(),
		kubernetes:  k,
	}
}

func (i *InitTokenPolicy) CreateToken() error {
	writeData := &vault.TokenCreateRequest{
		ID:          i.initToken,
		DisplayName: i.policy_name + "-creator",
		TTL:         "8760h",
		Period:      "8760h",
		Policies:    []string{i.policy_name + "-creator"},
	}

	_, err := i.kubernetes.vaultClient.Auth().Token().CreateOrphan(writeData)
	if err != nil {
		logrus.Fatal("Failed to create init token", err)
	}
	logrus.Infof("Created init token %s ", i.policy_name)

	return nil

}

func (i *InitTokenPolicy) WriteInitToken() error {

	path := filepath.Join(i.kubernetes.clusterID, "secrets", "init-token-"+i.role_name)
	writeData := map[string]interface{}{
		"init_token": i.initToken,
	}

	_, err := i.kubernetes.vaultClient.Logical().Write(path, writeData)
	if err != nil {
		logrus.Fatal("Failed to create init token", err)
	}
	logrus.Infof("Written init token %s ", i.policy_name)

	return nil
}
