package cert

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

type Cert struct {
	role        string
	commonName  string
	destination string
	bitSize     int
	keyType     string
	ipSans      []string
	sanHosts    []string
	owner       string
	group       string

	vaultClient *vault.Client
	Log         *logrus.Entry
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

func (c *Cert) RunCert() error {
	if err := c.ensureDestination(); err != nil {
		return fmt.Errorf("Error ensuring destination:\n%s", err)
	}

	return nil
}

// Ensure destination path is a directory
func (c *Cert) ensureDestination() error {
	fi, err := os.Stat(c.Destination())

	// Path exists but throws an error
	if err != nil && os.IsExist(err) {
		return fmt.Errorf("Error trying to read at location '%s'\n%s", c.Destination(), err)
	}

	// Path doesn't exist
	if err != nil && os.IsNotExist(err) {
		os.Mkdir(c.Destination(), 0755)
		c.Log.Debugf("Destination directory doesn't exist. Directory created: %s", c.Destination())
		return nil
	}

	// Exists but is not a directory
	if mode := fi.Mode(); !mode.IsDir() {
		return fmt.Errorf("Destination '%s' is not a directory", c.Destination())
	}

	c.Log.Debugf("Destination directory exists")

	return nil
}

// Ensure -key.pem exists, and has correct size and key type
func (c *Cert) ensureKey() error {
	path := filepath.Join(c.Destination(), "-key.pem")
	fi, err := os.Stat(path)

	// Path doesn't exist
	if err != nil && os.IsNotExist(err) {
		c.Log.Debugf("Key doesn't exist at path: %s", path)
		return nil
	}

	return nil
}

func (c *Cert) generateKey() error {

	return nil
}

func (c *Cert) deleteFile(path string) error {

	return nil
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
