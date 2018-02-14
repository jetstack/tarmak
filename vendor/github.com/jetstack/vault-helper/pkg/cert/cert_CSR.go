package cert

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

func (c *Cert) RequestCertificate() error {
	if err := c.verifyCertificates(); err != nil {
		c.Log.Debugf("Couldn't verify certificates: %v", err)
		c.Log.Info("Generating new certificates")
		return c.createNewCerts()
	}

	c.Log.Infof("Found certificates at %s", c.Destination())
	c.Log.Info("Certificates verified.")

	return nil
}

func (c *Cert) createNewCerts() error {
	ipSans := strings.Join(c.IPSans(), ",")
	hosts := strings.Join(c.SanHosts(), ",")

	path := filepath.Clean(c.Role())

	data := map[string]interface{}{
		"common_name": c.CommonName(),
		"ip_sans":     ipSans,
		"alt_names":   hosts,
	}
	sec, err := c.writeCSR(path, data)
	if err != nil {
		return fmt.Errorf("error writing CSR to vault at '%s': %v", path, err)
	}

	cert, certCA, err := c.decodeSec(sec)
	if err != nil {
		return fmt.Errorf("failed to decode secret from CSR: %v", err)
	}

	if cert == "" {
		return errors.New("no certificate received")
	}
	if certCA == "" {
		return errors.New("no ca certificate received")
	}

	c.Log.Infof("New certificate received for: %s", c.CommonName())

	certPath := filepath.Clean(c.Destination() + ".pem")
	caPath := filepath.Clean(c.Destination() + "-ca.pem")

	if err := c.storeCertificate(certPath, cert); err != nil {
		return fmt.Errorf("error storing certificate at path '%s': %v", certPath, err)
	}
	if err := c.storeCertificate(caPath, certCA); err != nil {
		return fmt.Errorf("error storing ca certificate at path '%s': %v", caPath, err)
	}

	return nil
}

func (c *Cert) checkExistingCerts(path string) (exist bool, err error) {
	fi, err := os.Stat(path)

	// Path exists but throws an error
	if err != nil && os.IsExist(err) {
		return true, fmt.Errorf("failed to read file at location '%s': %v", path, err)
	}

	// Path doesn't exist
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	// Exists but is a directory
	if mode := fi.Mode(); mode.IsDir() {
		return true, fmt.Errorf("destination '%s' is a directory", path)
	}

	return true, nil
}

func (c *Cert) verifyCertificates() error {
	conf := vault.DefaultConfig()
	conf.Address = c.InstanceToken().VaultClient().Address()

	tConf := &vault.TLSConfig{
		CAPath:     filepath.Clean(filepath.Dir(c.Destination())),
		CACert:     filepath.Clean(c.Destination() + "-ca.pem"),
		ClientCert: filepath.Clean(c.Destination() + ".pem"),
		ClientKey:  filepath.Clean(c.Destination() + "-key.pem"),
	}

	if err := conf.ConfigureTLS(tConf); err != nil {
		return fmt.Errorf("error verifying cert: %v", err)
	}

	return nil
}

func (c *Cert) decodeSec(sec *vault.Secret) (cert string, certCA string, err error) {
	if sec == nil {
		return "", "", errors.New("no secret returned from vault")
	}

	certField, ok := sec.Data["certificate"]
	if !ok {
		return "", "", errors.New("certificate field not found")
	}

	cert, ok = certField.(string)
	if !ok {
		return "", "", errors.New("failed to convert certificiate field to string")
	}

	if certCAField, ok := sec.Data["ca_chain"]; ok {
		certCA, ok = certCAField.(string)
		if !ok {
			return "", "", errors.New("failed to convert ca chain certificiate field to string")
		}
	} else {
		c.Log.Debugf("CA chain field not found - trying issuing CA")
		certCAField, ok := sec.Data["issuing_ca"]
		if !ok {
			return "", "", errors.New("issuing ca certificate or ca chain certificate field not found")
		}
		certCA, ok = certCAField.(string)
		if !ok {
			return "", "", errors.New("failed to convert issuing ca certificiate field to string")
		}
	}

	return cert, certCA, err
}

func (c *Cert) createCSR() (csr []byte, err error) {
	names := pkix.Name{
		CommonName: c.CommonName(),
		Organization: c.Organisation(),
	}
	var csrTemplate = x509.CertificateRequest{
		Subject:            names,
		SignatureAlgorithm: x509.SHA512WithRSA,
	}

	key, err := x509.ParsePKCS1PrivateKey(c.Data().Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key bytes: %v", err)
	}

	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %v", err)
	}

	csr = pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	return csr, nil
}

func (c *Cert) writeCSR(path string, data map[string]interface{}) (secret *vault.Secret, err error) {
	csr, err := c.createCSR()
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %v", err)
	}

	pemBytes := []byte(csr)
	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, fmt.Errorf("CSR contains no data: %v", err)
	}
	data["csr"] = string(csr)

	return c.InstanceToken().VaultClient().Logical().Write(path, data)
}

func (c *Cert) storeCertificate(path, cert string) error {
	fi, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %v", path, err)
	}
	defer fi.Close()

	if _, err := fi.Write([]byte(cert)); err != nil {
		return fmt.Errorf("failed to write certificate to file '%s': %v", path, err)
	}

	if err := c.WritePermissions(path, os.FileMode(0644)); err != nil {
		return fmt.Errorf("failed to set permissons of certificate file '%s': %s", path, err)
	}

	c.Log.Infof("Certificate written to: %s", path)

	return nil
}
