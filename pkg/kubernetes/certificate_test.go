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

func TestKubeletOrganization(t *testing.T) {

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

	data := map[string]interface{}{
		"common_name": "kubelet",
	}

	path := filepath.Join("test-cluster", "pki", "k8s", "sign", "kubelet")

	sec, err := writeCertificate(path, data, vault.Client())

	if err != nil {
		t.Errorf("Error reading signiture: ", err)
		return
	}

	cert_ca := sec.Data["certificate"].(string)

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(cert_ca))
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
		logrus.Errorf("%s", err)
		return
	}

	for _, org := range cert.Subject.Organization {
		if org == "system:nodes" || org == "system:masters" {
			logrus.Infof("%s in kubelet subject", org)
		} else {
			t.Errorf("%s shouldn't be in subject", org)
			return
		}
	}

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

	for _, role := range []string{"server", "client"} {
		verify_certificate(t, "etcd-k8s", role, vault.Client())
	}

	for _, role := range []string{"kube-scheduler", "kube-apiserver", "kube-controller-manager", "kube-proxy", "admin"} {
		verify_certificate(t, "k8s", role, vault.Client())
	}

}

func TestApiServerCanAdd(t *testing.T) {
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

	data := map[string]interface{}{
		"common_name": "kube-apiserver",
		"alt_names":   "THISisAny,HoSTName",
		"ip_sans":     "245.32.41.23,0.0.0.0,255.255.255.255",
	}

	path := filepath.Join("test-cluster", "pki", "k8s", "sign", "kube-apiserver")

	_, err := writeCertificate(path, data, vault.Client())
	if err != nil {
		t.Errorf("Error writting signiture: ", err)
		return
	}

	logrus.Infof("Can add any ip/hostname to apiserver")

}

func writeCertificate(path string, data map[string]interface{}, vault *vault.Client) (*vault.Secret, error) {

	pkixName := pkix.Name{
		Country:            []string{"Earth"},
		Organization:       []string{"Mother Nature"},
		OrganizationalUnit: []string{"Solar System"},
		Locality:           []string{""},
		Province:           []string{""},
		StreetAddress:      []string{"Solar System"},
		PostalCode:         []string{"Planet # 3"},
		SerialNumber:       "123",
		CommonName:         "foo.com",
	}

	csr, err := createCertificateAuthority(pkixName, time.Hour*60, 2048)
	if err != nil {
		return nil, fmt.Errorf("Error generating certificate: ", err)
	}

	data["csr"] = string(csr)

	pemBytes := []byte(csr)
	pemBlock, pemBytes := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, fmt.Errorf("csr contain no data", err)
	}

	sec, err := vault.Logical().Write(path, data)
	return sec, err
}

func verify_certificate(t *testing.T, name, role string, vault *vault.Client) {

	data := map[string]interface{}{
		"common_name": role,
		"alt_names":   "etcd-1.tarmak.local,localhost",
		"ip_sans":     "127.0.0.1",
	}

	path := filepath.Join("test-cluster", "pki", name, "sign", role)

	sec, err := writeCertificate(path, data, vault)

	if err != nil {
		t.Errorf("Error writting signiture: ", err)
		return
	}

	cert_ca := sec.Data["certificate"].(string)
	//issue_ca := sec.Data["issuing_ca"].(string)

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(cert_ca))
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

	logrus.Infof("%s - %s", role, cert.IPAddresses)

	opts := x509.VerifyOptions{
		DNSName: "127.0.0.1",
		Roots:   roots,
	}
	_, err = cert.Verify(opts)
	if err != nil && (role == "server" || role == "kube-apiserver") {
		t.Error("failed to verify certificate: " + err.Error())
		return
	}
	if role == "server" || role == "kube-apiserver" {
		logrus.Infof("127.0.0.1 in certificate %s - %s", name, role)
	}

	opts.DNSName = "etcd-1.tarmak.local"
	_, err = cert.Verify(opts)
	if err != nil && (role == "server" || role == "kube-apiserver") {
		t.Error("failed to verify certificate: " + err.Error())
		return
	}
	if role == "server" || role == "kube-apiserver" {
		logrus.Infof("etcd-1.tarmak.local in certificate %s - %s", name, role)
	}

	opts.DNSName = "localhost"
	_, err = cert.Verify(opts)
	if err != nil && (role == "server" || role == "kube-apiserver") {
		t.Error("failed to verify certificate: " + err.Error())
		return
	}
	if role == "server" || role == "kube-apiserver" {
		logrus.Infof("localhost in certificate %s - %s", name, role)
	}

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