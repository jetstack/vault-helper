package kubernetes

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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

	mount, err := GetMountByPath(g.kubernetes.vaultClient, g.Path())
	if err != nil {
		return err
	}

	if mount == nil {
		logrus.Infof("No secrects mount found for: %s", g.Path())
		err = g.kubernetes.vaultClient.Sys().Mount(
			g.Path(),
			&vault.MountInput{
				Description: "Kubernetes " + g.kubernetes.clusterID + " secrets",
				Type:        "generic",
			},
		)

		if err != nil {
			return fmt.Errorf("error creating mount: %s", err)
		}

		logrus.Infof("Mounted secrets")
	}

	rsaKeyPath := filepath.Join(g.Path(), "service-accounts")
	if secret, err := g.kubernetes.vaultClient.Logical().Read(rsaKeyPath); err != nil {
		return fmt.Errorf("error checking for secret %s: %s", rsaKeyPath, err)
	} else if secret == nil {
		err = g.writeNewRSAKey(rsaKeyPath, 4096)
		if err != nil {
			return fmt.Errorf("error creating rsa key at %s: %s", rsaKeyPath, err)
		}
	}

	return nil
}

func (g *Generic) writeNewRSAKey(secretPath string, bitSize int) error {

	reader := rand.Reader
	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return fmt.Errorf("error generating rsa key: %s", err)
	}

	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	err = pem.Encode(writer, privateKey)
	if err != nil {
		return fmt.Errorf("error encoding rsa key in PEM: %s", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing buffer: %s", err)
	}

	writeData := map[string]interface{}{
		"key": buf.String(),
	}

	_, err = g.kubernetes.vaultClient.Logical().Write(secretPath, writeData)

	if err != nil {
		return fmt.Errorf("error writting key to secrets: %s", err)
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
		return fmt.Errorf("Failed to create init token: %s", err)
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
		return fmt.Errorf("Failed to create init token: %s", err)
	}
	logrus.Infof("Written init token %s ", i.policy_name)

	i.kubernetes.secretsGeneric.initTokens[i.role_name] = i.initToken

	return nil
}
