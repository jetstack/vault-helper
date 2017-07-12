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
	return string(uuID[:])
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
