package kubernetes

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"

	"github.com/jetstack/vault-helper/pkg/testing/vault_dev"
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

func TestOrganisations(t *testing.T) {
	vault := vault_dev.New()
	if err := vault.Start(); err != nil {
		t.Errorf("unable to initialise vault dev server for integration tests: %v", err)
		return
	}
	defer vault.Stop()

	k := New(vault.Client(), logrus.NewEntry(logrus.New()))
	k.SetClusterID("test-cluster")

	if err := k.Ensure(); err != nil {
		t.Errorf("error ensuring %v", err)
		return
	}

	for _, role := range []string{"kube-scheduler", "kube-apiserver", "kube-controller-manager", "kube-proxy"} {
		path := filepath.Join("test-cluster", "pki", "k8s", "sign", role)
		if err := OrgMatch(role, path, []string{""}, vault.Client()); err != nil {
			t.Error("error matching organisation: ", err)
			return
		}
	}

	for _, role := range []string{"server", "client"} {
		path := filepath.Join("test-cluster", "pki", "etcd-k8s", "sign", role)
		if err := OrgMatch(role, path, []string{""}, vault.Client()); err != nil {
			t.Errorf("error matching organisation: %v", err)
			return
		}
	}

	path := filepath.Join("test-cluster", "pki", "k8s", "sign", "admin")
	if err := OrgMatch("admin", path, []string{"system:masters"}, vault.Client()); err != nil {
		t.Errorf("error matching organisation: %v", err)
		return
	}

	path = filepath.Join("test-cluster", "pki", "k8s", "sign", "kubelet")
	if err := OrgMatch("kubelet", path, []string{"system:nodes"}, vault.Client()); err != nil {
		t.Errorf("error matching organisation: %v", err)
		return
	}

	return
}

func OrgMatch(role, path string, match []string, vaultClient *vault.Client) error {
	data := map[string]interface{}{
		"common_name": role,
	}

	sec, err := writeCertificate(path, data, vaultClient)
	if err != nil {
		return fmt.Errorf("failed to read signiture: %v", err)
	}

	cert_field, ok := sec.Data["certificate"]
	if !ok {
		return errors.New("certificate field not found")
	}
	cert_ca, ok := cert_field.(string)
	if !ok {
		return errors.New("certificate field not string")
	}

	roots := x509.NewCertPool()
	ok = roots.AppendCertsFromPEM([]byte(cert_ca))
	if !ok {
		return errors.New("failed to parse root certificate")
	}

	block, _ := pem.Decode([]byte(cert_ca))
	if block == nil {
		return errors.New("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	if len(match) > 0 && match[0] == "" {
		match = append(match[:0], match[1:]...)
	}

	matchcp := make([]string, len(match), cap(match))
	copy(matchcp, match)

	for _, org := range cert.Subject.Organization {
		matched := false
		for i, m := range match {
			if m == org {
				match = append(match[:i], match[i+1:]...)
				matched = true
				continue
			}
		}
		if !matched {
			return fmt.Errorf("error, unexpected organisation: exp:%s got:%s", cert.Subject.Organization, matchcp)
		}
	}

	if len(match) > 0 {
		return fmt.Errorf("Error, organisations not found but expected:%s got:%s", matchcp, cert.Subject.Organization)
	}

	logrus.Infof("%s - Expected organisations found: '%s'", role, matchcp)

	return nil
}

func TestCertificates(t *testing.T) {
	vault := vault_dev.New()
	if err := vault.Start(); err != nil {
		t.Errorf("unable to initialise vault dev server for integration tests: %v", err)
		return
	}
	defer vault.Stop()

	k := New(vault.Client(), logrus.NewEntry(logrus.New()))
	k.SetClusterID("test-cluster")

	if err := k.Ensure(); err != nil {
		t.Errorf("failed to ensure %v", err)
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
		t.Errorf("unable to initialise vault dev server for integration tests: %v", err)
		return
	}
	defer vault.Stop()

	k := New(vault.Client(), logrus.NewEntry(logrus.New()))
	k.SetClusterID("test-cluster")

	if err := k.Ensure(); err != nil {
		t.Errorf("failed to ensure %v", err)
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
		t.Errorf("error writting signiture: %v", err)
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

	csr, err := createCertificateSigningRequest(pkixName, time.Hour*60, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %v", err)
	}

	data["csr"] = string(csr)

	pemBytes := []byte(csr)
	pemBlock, pemBytes := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, fmt.Errorf("csr contains no data: %v", err)
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
		t.Errorf("error writting signiture: %v", err)
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
		t.Errorf("failed to verify certificate: %v", err.Error())
		return
	}
	if role == "server" || role == "kube-apiserver" {
		logrus.Infof("etcd-1.tarmak.local in certificate %s - %s", name, role)
	}

	opts.DNSName = "localhost"
	_, err = cert.Verify(opts)
	if err != nil && (role == "server" || role == "kube-apiserver") {
		t.Errorf("failed to verify certificate: %v", err.Error())
		return
	}
	if role == "server" || role == "kube-apiserver" {
		logrus.Infof("localhost in certificate %s - %s", name, role)
	}

}

func createCertificateSigningRequest(names pkix.Name, expiration time.Duration, size int) ([]byte, error) {
	// step: generate a keypair
	keys, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return nil, fmt.Errorf("unable to genarate private keys, error: %v", err)
	}

	// step: generate a csr template
	var csrTemplate = x509.CertificateRequest{
		Subject:            names,
		SignatureAlgorithm: x509.SHA512WithRSA,
	}

	// step: generate the csr request
	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)
	if err != nil {
		return nil, err
	}

	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	return csr, nil
}
