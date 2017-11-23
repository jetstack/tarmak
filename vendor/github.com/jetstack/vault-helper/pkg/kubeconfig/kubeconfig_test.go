package kubeconfig

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/jetstack/vault-helper/pkg/cert"
	"github.com/jetstack/vault-helper/pkg/instanceToken"
	"github.com/jetstack/vault-helper/pkg/kubernetes"
	"github.com/jetstack/vault-helper/pkg/testing/vault_dev"
	"gopkg.in/yaml.v2"
)

var vaultDev *vault_dev.VaultDev

var tempDirs []string

func TestMain(m *testing.M) {
	vaultDev = initVaultDev()

	// this runs all tests
	returnCode := m.Run()

	// shutdown vault
	vaultDev.Stop()

	// clean up tempdirs
	for _, dir := range tempDirs {
		os.RemoveAll(dir)
	}

	// return exit code according to the test runs
	os.Exit(returnCode)
}

func TestKubeconf_Busy_Vault(t *testing.T) {
	initKubernetes(t, vaultDev)
	c := initCert(t, vaultDev)

	if err := c.InstanceToken().WriteTokenFile(c.InstanceToken().InitTokenFilePath(), vault_dev.RootTokenDev); err != nil {
		t.Fatalf("error setting token for test: %v", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("error runinning cert: %v", err)
	}

	u := initKubeconf(t, c)

	if err := u.RunKube(); err != nil {
		t.Fatalf("error runinning kubeconfig: %v", err)
	}

	yml := importYaml(t, u.FilePath())

	ymlKeyBef := yml.Users[0].User.ClientKeyData
	ymlCerBef := yml.Users[0].User.ClientCertificateData
	ymlCABef := yml.Clusters[0].Cluster.CertificateAuthorityData

	u.Log.Infof("-- Second run call --")
	u.Cert().InstanceToken().VaultClient().SetToken("foo-bar")
	defer u.Cert().InstanceToken().VaultClient().SetToken(vault_dev.RootTokenDev)
	if err := u.RunKube(); err != nil {
		t.Fatalf("Expected 400 error, premisson denied")
	}

	yml = importYaml(t, u.FilePath())

	ymlKeyAft := yml.Users[0].User.ClientKeyData
	ymlCerAft := yml.Users[0].User.ClientCertificateData
	ymlCAAft := yml.Clusters[0].Cluster.CertificateAuthorityData

	if ymlKeyBef != ymlKeyAft {
		t.Fatalf("Yaml key data before does not match yaml key data after. expected no changed. bef=%v aft=%v", ymlKeyBef, ymlKeyAft)
	}
	if ymlCABef != ymlCAAft {
		t.Fatalf("Yaml CA data before does not match yaml CA data after. expected no changed. bef=%v aft=%v", ymlCABef, ymlCAAft)
	}
	if ymlCerBef != ymlCerAft {
		t.Fatalf("Yaml cert data before does not match yaml cert data after. expected no changed. bef=%v aft=%v", ymlCerBef, ymlCerAft)
	}

}

// Test permissons of created files
func TestKubeconf_File_Perms(t *testing.T) {

	initKubernetes(t, vaultDev)
	c := initCert(t, vaultDev)

	if err := c.InstanceToken().WriteTokenFile(c.InstanceToken().InitTokenFilePath(), vault_dev.RootTokenDev); err != nil {
		t.Fatalf("error setting token for test: %v", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("error runinning cert: %v", err)
	}

	u := initKubeconf(t, c)

	if err := u.RunKube(); err != nil {
		t.Fatalf("error runinning kubeconfig: %v", err)
	}

	yaml := filepath.Clean(u.FilePath())
	checkFilePerm(t, yaml, os.FileMode(0600))
	checkOwnerGroup(t, yaml)
}

func TestKubeconf_Cert_Data(t *testing.T) {
	initKubernetes(t, vaultDev)
	c := initCert(t, vaultDev)

	if err := c.InstanceToken().WriteTokenFile(c.InstanceToken().InitTokenFilePath(), vault_dev.RootTokenDev); err != nil {
		t.Fatalf("error setting token for test: %v", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("error runinning cert: %v", err)
	}

	u := initKubeconf(t, c)

	if err := u.RunKube(); err != nil {
		t.Fatalf("error runinning kubeconfig: %v", err)
	}

	keyPem := filepath.Clean(c.Destination() + "-key.pem")
	cerPem := filepath.Clean(c.Destination() + ".pem")
	caPem := filepath.Clean(c.Destination() + "-ca.pem")

	key, err := u.encode64File(keyPem)
	if err != nil {
		t.Fatalf("failed to encode data at file '%s': %v", keyPem, err)
	}
	ca, _ := u.encode64File(caPem)
	if err != nil {
		t.Fatalf("failed to encode data at file '%s': %v", caPem, err)
	}
	cer, _ := u.encode64File(cerPem)
	if err != nil {
		t.Fatalf("failed to encode data at file '%s': %v", cerPem, err)
	}

	yml := importYaml(t, u.FilePath())

	ymlKey := yml.Users[0].User.ClientKeyData
	ymlCer := yml.Users[0].User.ClientCertificateData
	ymlCA := yml.Clusters[0].Cluster.CertificateAuthorityData

	if key != ymlKey {
		t.Fatalf("key data and file key data do not match. exp=%v got=%v", key, ymlKey)
	}
	if cer != ymlCer {
		t.Fatalf("Cert data and file Cert data do not match. exp=%v got=%v", ca, ymlCer)
	}
	if ca != ymlCA {
		t.Fatalf("CA data and file CA data do not match. exp=%v got=%v", ca, ymlCA)
	}

}

func importYaml(t *testing.T, path string) (yml *KubeY) {

	data := getFileData(t, path)

	err := yaml.Unmarshal([]byte(data), &yml)
	if err != nil {
		t.Fatalf("failed to unmarshal yaml file data: %v", err)
	}

	return yml
}

func getFileData(t *testing.T, path string) (data []byte) {

	fi, err := os.Open(path)
	if err != nil {
		t.Fatalf("unexpected error reading file '%s': %v", path, err)
	}

	fileinfo, err := fi.Stat()
	if err != nil {
		t.Fatalf("unable to get file info '%s': %v", path, err)
	}

	size := fileinfo.Size()
	bytes := make([]byte, size)

	buffer := bufio.NewReader(fi)
	_, err = buffer.Read(bytes)
	if err != nil {
		t.Fatalf("unable to read bytes from file '%s': %v", path, err)
	}

	return bytes
}

// Check permissions of a file
func checkFilePerm(t *testing.T, path string, mode os.FileMode) {
	if fi, err := os.Stat(path); err != nil {
		t.Fatalf("error finding stats of '%s': %v", path, err)
	} else if fi.IsDir() {
		t.Fatalf("file should not be directory %s", path)
	} else if perm := fi.Mode(); perm != mode {
		t.Fatalf("destination has incorrect file permissons. exp=%s got=%s", mode, perm)
	}
}

// Check permissions of a file
func checkOwnerGroup(t *testing.T, path string) {
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("error finding stats of '%s': %v", path, err)
	}

	curr, err := user.Current()
	if err != nil {
		t.Fatalf("error retrieving current user info: %v", curr)
	}

	uid := fmt.Sprint(fi.Sys().(*syscall.Stat_t).Uid)
	gid := fmt.Sprint(fi.Sys().(*syscall.Stat_t).Gid)

	if uid != curr.Uid {
		t.Fatalf("file uid '%s' doesn't match user '%s' at %s", uid, curr.Uid, path)
	} else if gid != curr.Gid {
		t.Fatalf("file gid '%s' doesn't match user group '%s' at %s", gid, curr.Gid, path)
	}
}

// Init kubernetes for testing
func initKubernetes(t *testing.T, vaultDev *vault_dev.VaultDev) *kubernetes.Kubernetes {
	k := kubernetes.New(vaultDev.Client(), logrus.NewEntry(logrus.New()))
	k.SetClusterID("test-cluster")

	if err := k.Ensure(); err != nil {
		t.Fatalf("failed to ensure kubernetes: %v", err)
	}

	return k
}

// Start vault_dev for testing
func initVaultDev() *vault_dev.VaultDev {
	vaultDev := vault_dev.New()

	if err := vaultDev.Start(); err != nil {
		logrus.Fatalf("unable to initialise vault dev server for integration tests: %v", err)
	}

	return vaultDev
}

// Init Cert for tesing
func initCert(t *testing.T, vaultDev *vault_dev.VaultDev) (c *cert.Cert) {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	log := logrus.NewEntry(logger)

	// setup temporary directory for tests
	dir, err := ioutil.TempDir("", "test-cluster-dir")
	if err != nil {
		t.Fatal(err)
	}
	tempDirs = append(tempDirs, dir)
	i := initInstanceToken(t, vaultDev, dir)

	c = cert.New(log, i)
	c.SetRole("test-cluster/pki/k8s/sign/kube-apiserver")
	c.SetCommonName("k8s")
	c.SetBitSize(2048)
	c.InstanceToken().SetVaultConfigPath(dir)
	c.SetDestination(dir + "/test")

	if usr, err := user.Current(); err != nil {
		t.Fatalf("error getting info on current user: %v", err)
	} else {
		c.SetOwner(usr.Username)
		c.SetGroup(usr.Username)
	}

	return c
}

// Init Kubeconfig for tesing
func initKubeconf(t *testing.T, cert *cert.Cert) (u *Kubeconfig) {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	log := logrus.NewEntry(logger)

	u = New(log, cert)
	u.SetFilePath(cert.Destination())

	return u
}

// Init instance token for testing
func initInstanceToken(t *testing.T, vaultDev *vault_dev.VaultDev, dir string) *instanceToken.InstanceToken {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	log := logrus.NewEntry(logger)

	i := instanceToken.New(vaultDev.Client(), log)

	i.SetVaultConfigPath(dir)

	if _, err := os.Stat(i.InitTokenFilePath()); os.IsNotExist(err) {
		ifile, err := os.Create(i.InitTokenFilePath())
		if err != nil {
			t.Fatalf("%s", err)
		}
		defer ifile.Close()
	}

	_, err := os.Stat(i.TokenFilePath())
	if os.IsNotExist(err) {
		tfile, err := os.Create(i.TokenFilePath())
		if err != nil {
			t.Fatalf("%s", err)
		}
		defer tfile.Close()
	}

	i.WipeTokenFile(i.InitTokenFilePath())
	i.WipeTokenFile(i.TokenFilePath())

	return i
}
