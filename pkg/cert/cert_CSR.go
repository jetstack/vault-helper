package cert

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

func (c *Cert) RequestCertificate() error {

	ipSans := strings.Join(c.IPSans(), ",")
	hosts := strings.Join(c.SanHosts(), ",")

	data := map[string]interface{}{
		"common_name": c.CommonName(),
		"ip_sans":     ipSans,
		"alt_names":   hosts,
	}
	sec, err := c.writeCSR(c.Destination(), data)
	if err != nil {
		return fmt.Errorf("Error writing CSR to vault at '%s':\n%s", c.Destination(), err)
	}

	cert, err := c.decodeSec(sec)
	if err != nil {
		return fmt.Errorf("Error decoding secret from CSR:\n%s", err)
	}

	if cert == nil {
		return fmt.Errorf("No certificate received.")
	}

	c.Log.Infof("New certificate received for: %s", c.CommonName())
	c.Log.Debugf("%s", cert)

	return nil
}

func (c *Cert) decodeSec(sec *vault.Secret) (cert []byte, err error) {

	cert_field, ok := sec.Data["certificate"]
	if !ok {
		return nil, fmt.Errorf("Error, certificate field not found")
	}
	cert_ca, ok := cert_field.(string)
	if !ok {
		return nil, fmt.Errorf("Error converting certificiate field to string")
	}

	roots := x509.NewCertPool()
	ok = roots.AppendCertsFromPEM([]byte(cert_ca))
	if !ok {
		return nil, fmt.Errorf("failed to parse root certificate")
	}

	block, _ := pem.Decode([]byte(cert_ca))
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}

	resp, err := x509.ParseCertificate(block.Bytes)

	return resp.Raw, err
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
