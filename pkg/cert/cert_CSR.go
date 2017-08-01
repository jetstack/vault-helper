package cert

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

func (c *Cert) RequestCertificate() error {

	ipSans := strings.Join(c.IPSans(), ",")
	hosts := strings.Join(c.SanHosts(), ",")

	path := filepath.Join(c.Role())

	data := map[string]interface{}{
		"common_name": c.CommonName(),
		"ip_sans":     ipSans,
		"alt_names":   hosts,
	}
	sec, err := c.writeCSR(path, data)
	if err != nil {
		return fmt.Errorf("Error writing CSR to vault at '%s':\n%s", c.Role(), err)
	}

	cert, certCA, err := c.decodeSec(sec)
	if err != nil {
		return fmt.Errorf("Error decoding secret from CSR:\n%s", err)
	}

	if cert == "" {
		return fmt.Errorf("No certificate received.")
	}
	if certCA == "" {
		return fmt.Errorf("No ca certificate received.")
	}

	c.Log.Infof("New certificate received for: %s", c.CommonName())

	certPath := filepath.Join(c.Destination(), ".pem")
	caPath := filepath.Join(c.Destination(), "-ca.pem")

	if err := c.storeCertificate(certPath, cert); err != nil {
		return fmt.Errorf("Error storing certificate at path '%s':\n%s", certPath, err)
	}
	if err := c.storeCertificate(caPath, certCA); err != nil {
		return fmt.Errorf("Error storing ca certificate at path '%s':\n%s", caPath, err)
	}

	return nil
}

func (c *Cert) decodeSec(sec *vault.Secret) (cert string, certCA string, err error) {

	if sec == nil {
		return "", "", fmt.Errorf("Error, no secret return from vault")
	}

	certCAField, ok := sec.Data["issuing_ca"]
	if !ok {
		return "", "", fmt.Errorf("Error, certificate field not found")
	}
	certCA, ok = certCAField.(string)
	if !ok {
		return "", "", fmt.Errorf("Error converting certificiate field to string")
	}

	certField, ok := sec.Data["certificate"]
	if !ok {
		return "", "", fmt.Errorf("Error, certificate field not found")
	}
	cert, ok = certField.(string)
	if !ok {
		return "", "", fmt.Errorf("Error converting certificiate field to string")
	}

	return cert, certCA, err
}

func (c *Cert) createCSR() (csr []byte, err error) {
	names := pkix.Name{
		CommonName: c.CommonName(),
	}
	var csrTemplate = x509.CertificateRequest{
		Subject:            names,
		SignatureAlgorithm: x509.SHA512WithRSA,
	}

	key, err := x509.ParsePKCS1PrivateKey(c.Data().Bytes)
	if err != nil {
		return nil, fmt.Errorf("Error parsing private key bytes: \n%s", err)
	}

	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, key)
	if err != nil {
		return nil, fmt.Errorf("Error creating CSR: %s", err)
	}

	csr = pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	return csr, nil
}

func (c *Cert) writeCSR(path string, data map[string]interface{}) (secret *vault.Secret, err error) {

	csr, err := c.createCSR()
	if err != nil {
		return nil, fmt.Errorf("Error generating certificate: %v", err)
	}

	pemBytes := []byte(csr)
	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, fmt.Errorf("CSR contains no data: %v", err)
	}
	data["csr"] = string(csr)

	return c.vaultClient.Logical().Write(path, data)
}

func (c *Cert) storeCertificate(path, cert string) error {

	fi, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Error opening file '%s': \n%s", path, err)
	}
	defer fi.Close()

	if _, err := fi.Write([]byte(cert)); err != nil {
		return fmt.Errorf("Error writting certificate to file '%s':\n%s", path, err)
	}

	if err := c.WritePermissions(path, 0644); err != nil {
		return fmt.Errorf("Error setting permissons of certificate file '%s':\n%s", path, err)
	}

	c.Log.Infof("Certificate written to: %s", path)

	return nil
}
