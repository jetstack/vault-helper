package cert

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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
	if err := c.ensureKey(); err != nil {
		return fmt.Errorf("Error ensuring key:\n%s", err)
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
	if err := c.ensureDestination(); err != nil {
		return fmt.Errorf("Error ensuring destination\n%s", err)
	}

	path := filepath.Join(c.Destination(), "-key.pem")
	_, err := os.Stat(path)

	// Path doesn't exist
	if err != nil && os.IsNotExist(err) {
		c.Log.Debugf("Pem file doesn't exist")
		c.Log.Infof("Key doesn't exist at path: %s", path)
		return c.genAndWriteKey(path)
	}

	//Path Exists
	c.Log.Debugf("Pem file exists")
	if err := c.loadKeyFromFile(path); err != nil {
		return fmt.Errorf("Error loading rsa key from file '%s':\n%s", path, err)
	}

	if c.KeyType() != c.Data().Type {
		c.Log.Infof("Key doesn't match expected type at path '%s'. exp=%s got=%s", path, c.KeyType(), c.Data().Type)
		// Wrong key type
		// Delete File, Generate new and write to file
		if err := c.deleteFile(path); err != nil {
			return err
		}
		return c.genAndWriteKey(path)
	}
	if c.BitSize() != c.PemSize() {
		c.Log.Infof("Key doesn't match expected size at path '%s'. exp=%d got=%d", path, c.BitSize(), c.PemSize())
		//Wrong bit size
		// Delete file, generate new and write to file
		if err := c.deleteFile(path); err != nil {
			return err
		}
		return c.genAndWriteKey(path)
	}
	c.Log.Debugf("No changes to Pem file")

	return nil
}

//Generate new key and write to file
func (c *Cert) genAndWriteKey(path string) error {
	c.Log.Infof("Generating new RSA key")
	if err := c.generateKey(path); err != nil {
		return fmt.Errorf("Error generating key:\n%s", err)
	}

	c.Log.Infof("Writing new key to file: %s", path)
	if err := c.writeKeyToFile(path); err != nil {
		return fmt.Errorf("Error saving key:\n%s", err)
	}

	return nil
}

func (c *Cert) loadKeyFromFile(path string) error {

	// Load PEM
	pemfile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Unable open file for reading '%s':\n%s", path, err)
	}

	// need to convert pemfile to []byte for decoding
	pemfileinfo, err := pemfile.Stat()
	if err != nil {
		return fmt.Errorf("Unable to get pem file info '%s':\n%s", path, err)
	}

	size := pemfileinfo.Size()
	pembytes := make([]byte, size)

	// read pemfile content into pembytes
	buffer := bufio.NewReader(pemfile)
	_, err = buffer.Read(pembytes)
	if err != nil {
		return fmt.Errorf("Unable to read pembyte from file:\n%s", err)
	}

	data, rest := pem.Decode([]byte(pembytes))
	if err != nil {
		return fmt.Errorf("Error decoding pem file. There was data in rest:\n%s", rest)
	}

	k, err := x509.ParsePKCS1PrivateKey(data.Bytes)
	if err != nil {
		return fmt.Errorf("Error parsing private key bytes: \n%s", err)
	}
	c.SetPemSize(k.D.BitLen())

	c.SetData(data)
	c.SetKeyType(data.Type)

	pemfile.Close()
	if err != nil {
		return fmt.Errorf("Unable to close pemfile:\n%s", err)
	}

	return nil
}

func (c *Cert) generateKey(path string) error {
	size := c.BitSize()
	key, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return fmt.Errorf("Error generating rsa key:\n%s", err)
	}

	key_bytes := x509.MarshalPKCS1PrivateKey(key)
	key_pem := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: key_bytes,
	}

	c.SetData(key_pem)

	return nil
}

// Save PEM file
func (c *Cert) writeKeyToFile(path string) error {
	pemfile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Error creating pem key file for writting:\n%s", err)
	}

	if err := pem.Encode(pemfile, c.Data()); err != nil {
		return fmt.Errorf("Error encoding key to pem file at'%s':\n%s", path, err)
	}

	if err := pemfile.Close(); err != nil {
		return fmt.Errorf("Error closing pem file at '%s':\n%s", path, err)
	}

	return nil
}

func (c *Cert) deleteFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("Error removing file at '%s':\n%s", path, err)
	}

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
