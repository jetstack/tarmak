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
		return fmt.Errorf("error ensuring destination: %v", err)
	}

	path := c.Destination() + "-key.pem"
	_, err := os.Stat(path)

	// Path doesn't exist
	if err != nil && os.IsNotExist(err) {
		c.Log.Debug("Pem file doesn't exist")
		c.Log.Infof("Key doesn't exist at path: %s", path)
		if err := c.genAndWriteKey(path); err != nil {
			return err
		}
		return c.WritePermissions(path, os.FileMode(0600))
	}

	//Path Exists
	c.Log.Debug("Pem file exists '-key.pem'")
	if err := c.loadKeyFromFile(path); err != nil {
		return fmt.Errorf("failed to load rsa key from file '%s': %v", path, err)
	}

	if c.KeyType() != c.Data().Type {
		c.Log.Warn("key doesn't match expected type at path '%s'. exp=%s got=%s", path, c.KeyType(), c.Data().Type)
		// Wrong key type
		// Delete File, Generate new and write to file
		if err := c.DeleteFile(path); err != nil {
			return err
		}
		return c.genAndWriteKey(path)
	}
	if c.BitSize() != c.PemSize() {
		c.Log.Infof("key doesn't match expected size at path '%s'. exp=%d got=%d", path, c.BitSize(), c.PemSize())
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
	dir := filepath.Dir(c.Destination())
	fi, err := os.Stat(dir)

	// Path exists but throws an error
	if err != nil && os.IsExist(err) {
		return fmt.Errorf("failed to read at location '%s': %v", dir, err)
	}

	// Path doesn't exist
	if err != nil && os.IsNotExist(err) {
		os.MkdirAll(dir, os.FileMode(0750))
		c.Log.Debugf("Destination directory doesn't exist. Directory created: %s", dir)
		return nil
	}

	// Exists but is not a directory
	if mode := fi.Mode(); !mode.IsDir() {
		return fmt.Errorf("destination '%s' is not a directory", dir)
	}

	if fi.Mode().Perm() != os.FileMode(0750) {
		c.Log.Debugf("Destination directory has incorrect permissons. Changing to 0770: %s", dir)
		c.WritePermissions(dir, os.FileMode(0750))
	}

	c.Log.Debugf("Destination directory exists")

	return nil
}

//Generate new key and write to file
func (c *Cert) genAndWriteKey(path string) error {
	c.Log.Infof("Generating new RSA key")
	if err := c.generateKey(); err != nil {
		return fmt.Errorf("error generating key: %v", err)
	}

	if err := c.writeKeyToFile(path); err != nil {
		return fmt.Errorf("error saving key to file '%s': %v", path, err)
	}
	c.Log.Infof("Key written to file: %s", path)

	return nil
}

func (c *Cert) loadKeyFromFile(path string) error {

	// Load PEM
	pemfile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to open file for reading '%s': %v", path, err)
	}

	// need to convert pemfile to []byte for decoding
	pemfileinfo, err := pemfile.Stat()
	if err != nil {
		return fmt.Errorf("unable to get pem file info '%s': %v", path, err)
	}

	size := pemfileinfo.Size()
	pembytes := make([]byte, size)

	// read pemfile content into pembytes
	buffer := bufio.NewReader(pemfile)
	_, err = buffer.Read(pembytes)
	if err != nil {
		return fmt.Errorf("unable to read pembyte from file: %v", err)
	}

	data, rest := pem.Decode([]byte(pembytes))
	if err != nil {
		return fmt.Errorf("failed to decode pem file. There was data left: %s", rest)
	}

	k, err := x509.ParsePKCS1PrivateKey(data.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key bytes: %v", err)
	}

	c.SetPemSize(k.N.BitLen())

	c.SetData(data)
	c.SetKeyType(data.Type)

	pemfile.Close()
	if err != nil {
		return fmt.Errorf("unable to close pemfile: %v", err)
	}

	return nil
}

func (c *Cert) generateKey() error {
	size := c.BitSize()
	key, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return fmt.Errorf("failed to generate rsa key: %v", err)
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
		return fmt.Errorf("failed to create pem key file for writting: %v", err)
	}

	if err := pem.Encode(pemfile, c.Data()); err != nil {
		return fmt.Errorf("failed to encode key to pem file at'%s': %v", path, err)
	}

	if err := pemfile.Close(); err != nil {
		return fmt.Errorf("error closing pem file at '%s': %v", path, err)
	}

	return nil
}
