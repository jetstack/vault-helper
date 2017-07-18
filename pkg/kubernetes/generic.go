package kubernetes

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	vault "github.com/hashicorp/vault/api"
)

type Generic struct {
	kubernetes *Kubernetes
	initTokens map[string]string
}

func (g *Generic) Ensure() error {
	err := g.GenerateSecretsMount()
	return err
}

func (g *Generic) Path() string {
	return filepath.Join(g.kubernetes.Path(), "generic")
}

func randomUUID() string {
	uuID := uuid.New()
	return uuID.String()
}

func (g *Generic) GenerateSecretsMount() error {

	secrets_path := filepath.Join(g.kubernetes.clusterID, "secrets")

	mount, err := GetMountByPath(g.kubernetes.vaultClient, secrets_path)
	if err != nil {
		return err
	}

	if mount == nil {
		logrus.Infof("No secrects mount found for: %s", secrets_path)
		err = g.kubernetes.vaultClient.Sys().Mount(
			secrets_path,
			&vault.MountInput{
				Description: "Kubernetes " + g.kubernetes.clusterID + " secrets",
				Type:        "generic",
			},
		)

		if err != nil {
			return fmt.Errorf("error creating mount: %s", err)
		}

		logrus.Infof("Mounted secrets")

		err = g.writeKey(secrets_path)
		if err != nil {
			return fmt.Errorf("error creating mount: %s", err)
		}

	} else {
		logrus.Infof("Secrets already mounted: %s", secrets_path)
	}

	return nil
}

func (g *Generic) writeKey(secrets_path string) error {

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

	_, err = g.kubernetes.vaultClient.Logical().Write(secrets_path, writeData)

	if err != nil {
		logrus.Fatal("Error writting key to secrets", err)
	}
	logrus.Infof("Key written to secrets")

	return nil
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

	i.kubernetes.secretsGeneric.initTokens[i.role_name] = i.initToken

	return nil
}
