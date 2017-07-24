package kubernetes

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

type caCertificate struct {
	privateKey string
	publicKey  string
	csr        string
}

type basicConstraints struct {
	IsCA       bool `asn1:"optional"`
	MaxPathLen int  `asn1:"optional,default:-1"`
}

func TestCertificates(t *testing.T) {
	vault := vault_dev.New()
	if err := vault.Start(); err != nil {
		t.Skip("unable to initialise vault dev server for integration tests: ", err)
		return
	}
	defer vault.Stop()

	k := New(vault.Client())
	k.SetClusterID("test-cluster")

	if err := k.Ensure(); err != nil {
		t.Errorf("Error ensuring %s", err)
		return
	}

	for _, role := range []string{"server"} {
		verify_certificate(t, "etcd-k8s", role, vault.Client())
	}

	for _, role := range []string{"kube-apiserver", "kube-scheduler", "kube-controller-manager", "kube-proxy", "admin"} {
		verify_certificate(t, "k8s", role, vault.Client())
	}

}

func verify_certificate(t *testing.T, name, role string, vault *vault.Client) {

	pkixName := pkix.Name{
		Country:            []string{"Earth"},
		Organization:       []string{"Mother Nature"},
		OrganizationalUnit: []string{"Solar System"},
		Locality:           []string{""},
		Province:           []string{""},
		StreetAddress:      []string{"Solar System"},
		PostalCode:         []string{"Planet # 3"},
		SerialNumber:       "123",
		CommonName:         "test-cluster",
	}

	csr, err := createCertificateAuthority(pkixName, time.Hour*60, 2048)
	if err != nil {
		t.Errorf("Error generating certificate: ", err)
		return
	}

	pemBytes := []byte(csr)
	pemBlock, pemBytes := pem.Decode(pemBytes)
	if pemBlock == nil {
		t.Errorf("csr contain no data", err)
		return
	}

	data := map[string]interface{}{
		"csr":         string(csr),
		"common_name": role,
		"alt_names":   "etcd-1.tarmak.local,localhost",
		"ip_sans":     "127.0.0.1",
	}

	path := filepath.Join("test-cluster", "pki", name, "sign", role)

	sec, err := vault.Logical().Write(path, data)

	if err != nil {
		t.Errorf("Error writting signiture: ", err)
		return
	}

	cert_ca := sec.Data["certificate"].(string)
	issue_ca := sec.Data["issuing_ca"].(string)

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(issue_ca))
	if !ok {
		t.Error("failed to parse root certificate")
		return
	}

	block, _ := pem.Decode([]byte(cert_ca))
	if block == nil {
		t.Error("failed to parse certificate PEM")
		return
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Error("failed to parse certificate: " + err.Error())
		return
	}

	opts := x509.VerifyOptions{
		DNSName: "127.0.0.1",
		Roots:   roots,
	}
	_, err = cert.Verify(opts)
	if err != nil {
		t.Error("failed to verify certificate: " + err.Error())
		return
	}
	logrus.Infof("127.0.0.1 in certificate %s - %s", name, role)

	opts.DNSName = "localhost"
	_, err = cert.Verify(opts)
	if err != nil {
		t.Error("failed to verify certificate: " + err.Error())
		return
	}
	logrus.Infof("localhost in certificate %s - %s", name, role)

	opts.DNSName = "etcd-1.tarmak.local"
	_, err = cert.Verify(opts)
	if err != nil {
		t.Error("failed to verify certificate: " + err.Error())
		return
	}
	logrus.Infof("ctcd-1.tarmak.local in certificate %s - %s", name, role)

}

//
// createCertificateAuthority generates a certificate authority request ready to be signed
//
func createCertificateAuthority(names pkix.Name, expiration time.Duration, size int) ([]byte, error) {
	// step: generate a keypair
	keys, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return nil, fmt.Errorf("unable to genarate private keys, error: %s", err)
	}

	val, err := asn1.Marshal(basicConstraints{true, 0})
	if err != nil {
		return nil, err
	}

	// step: generate a csr template
	var csrTemplate = x509.CertificateRequest{
		Subject:            names,
		SignatureAlgorithm: x509.SHA512WithRSA,
		ExtraExtensions: []pkix.Extension{
			{
				Id:       asn1.ObjectIdentifier{2, 5, 29, 19},
				Value:    val,
				Critical: true,
			},
		},
	}

	// step: generate the csr request
	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)
	if err != nil {
		logrus.Infof("%s", err)
		return nil, err
	}

	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	return csr, nil

}
