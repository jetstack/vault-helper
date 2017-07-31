package cert

import (
	"encoding/pem"
	"fmt"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

type Cert struct {
	role        string
	commonName  string
	destination string
	bitSize     int
	pemSize     int
	keyType     string
	ipSans      []string
	sanHosts    []string
	owner       string
	group       string
	data        *pem.Block

	vaultClient *vault.Client
	Log         *logrus.Entry
}

func (c *Cert) RunCert() error {
	if err := c.EnsureKey(); err != nil {
		return fmt.Errorf("Error ensuring key:\n%s", err)
	}

	if err := c.RequestCertificate(); err != nil {
		return fmt.Errorf("Error requesting certificate:\n%s", err)
	}
	return nil
}

func New(vaultClient *vault.Client, logger *logrus.Entry) *Cert {
	c := &Cert{
		role:        "",
		commonName:  "",
		destination: "",
		bitSize:     2048,
		keyType:     "RSA",
		ipSans:      []string{},
		sanHosts:    []string{},
		owner:       "",
		group:       "",
	}

	if vaultClient != nil {
		c.vaultClient = vaultClient
	}
	if logger != nil {
		c.Log = logger
	}

	return c
}

func (c *Cert) SetRole(role string) {
	c.role = role
}
func (c *Cert) Role() string {
	return c.role
}

func (c *Cert) SetCommonName(name string) {
	c.commonName = name
}
func (c *Cert) CommonName() string {
	return c.commonName
}

func (c *Cert) SetDestination(destination string) {
	c.destination = destination
}
func (c *Cert) Destination() string {
	return c.destination
}

func (c *Cert) SetBitSize(size int) {
	c.bitSize = size
}
func (c *Cert) BitSize() int {
	return c.bitSize
}

func (c *Cert) SetPemSize(size int) {
	c.pemSize = size
}
func (c *Cert) PemSize() int {
	return c.pemSize
}

func (c *Cert) SetKeyType(keyType string) {
	c.keyType = keyType
}
func (c *Cert) KeyType() string {
	return c.keyType
}

func (c *Cert) SetIPSans(ips []string) {
	c.ipSans = ips
}
func (c *Cert) IPSans() []string {
	return c.ipSans
}

func (c *Cert) SetSanHosts(hosts []string) {
	c.sanHosts = hosts
}
func (c *Cert) SanHosts() []string {
	return c.sanHosts
}

func (c *Cert) SetOwner(owner string) {
	c.owner = owner
}
func (c *Cert) Owner() string {
	return c.owner
}

func (c *Cert) SetGroup(group string) {
	c.group = group
}
func (c *Cert) Group() string {
	return c.group
}

func (c *Cert) SetData(data *pem.Block) {
	c.data = data
}
func (c *Cert) Data() *pem.Block {
	return c.data
}
