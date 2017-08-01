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
)

// Ensure -key.pem exists, and has correct size and key type
func (c *Cert) EnsureKey() error {
	if err := c.ensureDestination(); err != nil {
		return fmt.Errorf("Error ensuring destination\n%s", err)
	}

	path := filepath.Join(c.Destination(), "-key.pem")
	_, err := os.Stat(path)

	// Path doesn't exist
	if err != nil && os.IsNotExist(err) {
		c.Log.Debugf("Pem file doesn't exist")
		c.Log.Infof("Key doesn't exist at path: %s", path)
		if err := c.genAndWriteKey(path); err != nil {
			return err
		}
		return c.WritePermissions(path, os.FileMode(0600))
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
		if err := c.DeleteFile(path); err != nil {
			return err
		}
		return c.genAndWriteKey(path)
	}
	if c.BitSize() != c.PemSize() {
		c.Log.Infof("Key doesn't match expected size at path '%s'. exp=%d got=%d", path, c.BitSize(), c.PemSize())
		//Wrong bit size
		// Delete file, generate new and write to file
		if err := c.DeleteFile(path); err != nil {
			return err
		}
		return c.genAndWriteKey(path)
	}

	return c.WritePermissions(path, os.FileMode(0600))
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
		os.MkdirAll(c.Destination(), 0755)
		c.Log.Debugf("Destination directory doesn't exist. Directory created: %s", c.Destination())
		return nil
	}

	// Exists but is not a directory
	if mode := fi.Mode(); !mode.IsDir() {
		return fmt.Errorf("Destination '%s' is not a directory", c.Destination())
	}

	if fi.Mode().Perm().String() != "drwxr-xr-x" {
		c.Log.Debugf("Destination directory has incorrect permissons. Changing to 0775: %s", c.Destination())
		c.WritePermissions(c.Destination(), os.FileMode(0755))
	}

	c.Log.Debugf("Destination directory exists")

	return nil
}

//Generate new key and write to file
func (c *Cert) genAndWriteKey(path string) error {
	c.Log.Infof("Generating new RSA key")
	if err := c.generateKey(); err != nil {
		return fmt.Errorf("Error generating key:\n%s", err)
	}

	if err := c.writeKeyToFile(path); err != nil {
		return fmt.Errorf("Error saving key:\n%s", err)
	}
	c.Log.Infof("Key written to file: %s", path)

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

func (c *Cert) generateKey() error {
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
