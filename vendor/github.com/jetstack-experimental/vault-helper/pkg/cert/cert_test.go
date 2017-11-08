package cert

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/Sirupsen/logrus"

	"github.com/jetstack-experimental/vault-helper/pkg/instanceToken"
	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"
	"github.com/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

var vaultDev *vault_dev.VaultDev

var tempDirs []string

func TestMain(m *testing.M) {
	vaultDev = initVaultDev()
	initKubernetes(vaultDev)

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

// Test permissons of created files
func TestCert_File_Perms(t *testing.T) {

	c, i := initCert(t, vaultDev)

	if err := i.WriteTokenFile(i.InitTokenFilePath(), vault_dev.RootTokenDev); err != nil {
		t.Fatalf("error setting token for test: %v", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("error runinning cert: %v", err)
	}

	dir := filepath.Dir(c.Destination())
	if fi, err := os.Stat(dir); err != nil {
		t.Fatalf("error finding stats of '%s': %v", dir, err)
	} else if !fi.IsDir() {
		t.Fatalf("destination should be directory %s", dir)
	} else if perm := fi.Mode().Perm(); perm != os.FileMode(0750).Perm() {
		t.Fatalf("destination has incorrect file permissons. exp=0750 got=%d", perm)
	}

	curr, err := user.Current()
	if err != nil {
		t.Fatalf("error retrieving current user info: %v", curr)
	}

	keyPem := filepath.Clean(c.Destination() + "-key.pem")
	dotPem := filepath.Clean(c.Destination() + ".pem")
	caPem := filepath.Clean(c.Destination() + "-ca.pem")
	checkFilePerm(t, keyPem, os.FileMode(0600))
	checkOwnerGroup(t, keyPem, curr.Uid, curr.Gid)
	checkFilePerm(t, dotPem, os.FileMode(0644))
	checkOwnerGroup(t, dotPem, curr.Uid, curr.Gid)
	checkFilePerm(t, caPem, os.FileMode(0644))
	checkOwnerGroup(t, caPem, curr.Uid, curr.Gid)
}

// Test when passed int instead of string for owner/group
func TestCert_File_Perms_Int(t *testing.T) {
	c, i := initCert(t, vaultDev)

	curr, err := user.Current()
	if err != nil {
		t.Fatalf("error retrieving current user info: %v", curr)
	}

	c.SetGroup(curr.Uid)
	c.SetOwner(curr.Gid)

	if err := i.WriteTokenFile(i.InitTokenFilePath(), vault_dev.RootTokenDev); err != nil {
		t.Fatalf("error setting token for test: %v", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("error runinning cert: %v", err)
	}

	dir := filepath.Dir(c.Destination())
	if fi, err := os.Stat(dir); err != nil {
		t.Fatalf("error finding stats of '%s': %v", dir, err)
	} else if !fi.IsDir() {
		t.Fatalf("destination should be directory %s", dir)
	} else if perm := fi.Mode().Perm(); perm != os.FileMode(0750).Perm() {
		t.Fatalf("destination has incorrect file permissons. exp=0750 got=%d", perm)
	}

	keyPem := filepath.Clean(c.Destination() + "-key.pem")
	dotPem := filepath.Clean(c.Destination() + ".pem")
	caPem := filepath.Clean(c.Destination() + "-ca.pem")
	checkFilePerm(t, keyPem, os.FileMode(0600))
	checkOwnerGroup(t, keyPem, curr.Uid, curr.Gid)
	checkFilePerm(t, dotPem, os.FileMode(0644))
	checkOwnerGroup(t, dotPem, curr.Uid, curr.Gid)
	checkFilePerm(t, caPem, os.FileMode(0644))
	checkOwnerGroup(t, caPem, curr.Uid, curr.Gid)
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

// Check group and owner of a file
func checkOwnerGroup(t *testing.T, path string, uid, gid string) {
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("error finding stats of '%s': %v", path, err)
	}

	uidF := fmt.Sprint(fi.Sys().(*syscall.Stat_t).Uid)
	gidF := fmt.Sprint(fi.Sys().(*syscall.Stat_t).Gid)

	if uidF != uid {
		t.Fatalf("file uid '%s' doesn't match given uid '%s' at %s", uidF, uid, path)
	} else if gidF != gid {
		t.Fatalf("file gid '%s' doesn't match given group '%s' at %s", gidF, gid, path)
	}
}

// Verify CAs exist
func TestCert_Verify_CA(t *testing.T) {
	c, i := initCert(t, vaultDev)
	if err := i.WriteTokenFile(i.InitTokenFilePath(), vault_dev.RootTokenDev); err != nil {
		t.Fatalf("failed to set token for test: %v", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("error runinning cert: %s", err)
	}

	dotPem := filepath.Clean(c.Destination() + ".pem")
	dat, err := ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", dotPem, err)
	}
	if dat == nil {
		t.Fatalf("no certificate at file: '%s'", dotPem)
	}

	caPem := filepath.Clean(c.Destination() + "-ca.pem")
	dat, err = ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", caPem, err)
	}
	if dat == nil {
		t.Fatalf("no certificate at file '%s'. expected certificate", dotPem)
	}
}

// Test config file path
func TestCert_ConfigPath(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-cluster-dir")
	if err != nil {
		t.Fatal(err)
	}

	c, i := initCert(t, vaultDev)
	i.SetVaultConfigPath(dir)
	if err := i.WriteTokenFile(i.InitTokenFilePath(), vault_dev.RootTokenDev); err != nil {
		t.Fatalf("error setting token for test: %v", err)
	}

	dotPem := filepath.Clean(c.Destination() + ".pem")
	if _, err := os.Stat(dotPem); !os.IsNotExist(err) {
		t.Fatalf("expexted error 'File doesn't exist on file '.pem''. got: %v", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("error runinning cert: %v", err)
	}

	caPem := filepath.Clean(c.Destination() + "-ca.pem")
	if _, err := os.Stat(caPem); err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", caPem, err)
	}

	dat, err := ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", dotPem, err)
	}
	if dat == nil {
		t.Fatalf("no certificate at file '%s'", dotPem)
	}

	caPem = filepath.Clean(c.Destination() + "-ca.pem")
	dat, err = ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("failed to read certificate file path '%s': %v", caPem, err)
	}
	if dat == nil {
		t.Fatalf("no certificate at file '%s'", dotPem)
	}
}

// Test if already existing valid certificate and key, they are kept
func TestCert_Exist_NoChange(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-cluster-dir")
	if err != nil {
		t.Fatal(err)
	}

	c, i := initCert(t, vaultDev)
	i.SetVaultConfigPath(dir)
	if err := i.WriteTokenFile(i.InitTokenFilePath(), vault_dev.RootTokenDev); err != nil {
		t.Fatalf("failed to set token for test: %v", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("error running  cert: %v", err)
	}

	dotPem := filepath.Clean(c.Destination() + ".pem")
	datDotPem, err := ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", dotPem, err)
	}
	if datDotPem == nil {
		t.Fatalf("no certificate at file '%s'", dotPem)
	}

	caPem := filepath.Clean(c.Destination() + "-ca.pem")
	datCAPem, err := ioutil.ReadFile(caPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", caPem, err)
	}
	if datCAPem == nil {
		t.Fatalf("no certificate at file '%s'", dotPem)
	}

	keyPem := filepath.Clean(c.Destination() + "-key.pem")
	datKeyPem, err := ioutil.ReadFile(keyPem)
	if err != nil {
		t.Fatalf("error reading from key file path: '%s': %v", keyPem, err)
	}
	if datKeyPem == nil {
		t.Fatalf("no key at file '%s'", keyPem)
	}

	c.Log.Infof("-- Second run call --")

	if err := c.RunCert(); err != nil {
		if len(err.Error()) < 36 {
			t.Fatalf("unexpected error: %v", err)
		}
		str := "error renewing tokens: token not renewable: "
		errStr := err.Error()[:len(err.Error())-36]
		if errStr != str {
			t.Fatalf("unexpexted error. exp=%s got=%v", str, err)
		}
		c.Log.Infof("expected error: %v", err)
	}

	datDotPemAfter, err := ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", dotPem, err)
	}

	if string(datDotPem) != string(datDotPemAfter) {
		t.Fatalf("certificate has been changed after cert call even though it exists: %s", dotPem)
	}

	datCAPemAfter, err := ioutil.ReadFile(caPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", caPem, err)
	}
	if string(datCAPem) != string(datCAPemAfter) {
		t.Fatalf("certificate has been changed after cert call even though it exists: %s", caPem)
	}

	datKeyPemAfter, err := ioutil.ReadFile(keyPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", keyPem, err)
	}
	if string(datKeyPem) != string(datKeyPemAfter) {
		t.Fatalf("key has been changed after cert call even though it exists: %s", keyPem)
	}
}

func TestCert_Busy_Vault(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-cluster-dir")
	if err != nil {
		t.Fatal(err)
	}

	c, i := initCert(t, vaultDev)
	i.SetVaultConfigPath(dir)
	if err := i.WriteTokenFile(i.InitTokenFilePath(), vault_dev.RootTokenDev); err != nil {
		t.Fatalf("error setting token for test: %v", err)
	}

	if err := c.RunCert(); err != nil {
		t.Fatalf("error running  cert: %v", err)
	}

	dotPem := filepath.Clean(c.Destination() + ".pem")
	datDotPem, err := ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", dotPem, err)
	}
	if datDotPem == nil {
		t.Fatalf("no certificate at file '%s'", dotPem)
	}

	caPem := filepath.Clean(c.Destination() + "-ca.pem")
	datCAPem, err := ioutil.ReadFile(caPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", caPem, err)
	}
	if datCAPem == nil {
		t.Fatalf("no certificate at file '%s'", dotPem)
	}

	keyPem := filepath.Clean(c.Destination() + "-key.pem")
	datKeyPem, err := ioutil.ReadFile(keyPem)
	if err != nil {
		t.Fatalf("error reading from key file path: '%s': %v", keyPem, err)
	}
	if datKeyPem == nil {
		t.Fatalf("no key at file '%s'", keyPem)
	}

	c.Log.Infof("-- Second run call --")
	if err := c.InstanceToken().VaultClient().Sys().Seal(); err != nil {
		t.Fatalf("error sealing vault")
	}
	if err := c.InstanceToken().TokenRenewRun(); err == nil {
		t.Fatalf("expected 400 error, permission denied")
	}
	if err := c.RunCert(); err != nil {
		t.Fatalf("Unexpected error")
	}

	datDotPemAfter, err := ioutil.ReadFile(dotPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", dotPem, err)
	}

	if string(datDotPem) != string(datDotPemAfter) {
		t.Fatalf("certificate has been changed after cert call even though it exists: %s", dotPem)
	}

	datCAPemAfter, err := ioutil.ReadFile(caPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", caPem, err)
	}
	if string(datCAPem) != string(datCAPemAfter) {
		t.Fatalf("certificate has been changed after cert call even though it exists %s", caPem)
	}

	datKeyPemAfter, err := ioutil.ReadFile(keyPem)
	if err != nil {
		t.Fatalf("error reading from certificate file path: '%s': %v", keyPem, err)
	}
	if string(datKeyPem) != string(datKeyPemAfter) {
		t.Fatalf("key has been changed after cert call even though it exists %s", keyPem)
	}

}

// Init Cert for testing
func initCert(t *testing.T, vaultDev *vault_dev.VaultDev) (c *Cert, i *instanceToken.InstanceToken) {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	log := logrus.NewEntry(logger)

	// setup temporary directory for tests
	dir, err := ioutil.TempDir("", "test-cluster-dir")
	if err != nil {
		t.Fatal(err)
	}
	tempDirs = append(tempDirs, dir)

	i = initInstanceToken(t, vaultDev, dir)

	c = New(log, i)
	c.SetRole("test-cluster/pki/k8s/sign/kube-apiserver")
	c.SetCommonName("k8s")
	c.SetBitSize(2048)

	if usr, err := user.Current(); err != nil {
		t.Fatalf("error getting info on current user: %v", err)
	} else {
		c.SetOwner(usr.Username)
		c.SetGroup(usr.Username)
	}

	c.InstanceToken().SetVaultConfigPath(dir)
	c.SetDestination(dir + "/test")

	return c, i
}

// Init kubernetes for testing
func initKubernetes(vaultDev *vault_dev.VaultDev) *kubernetes.Kubernetes {
	k := kubernetes.New(vaultDev.Client(), logrus.NewEntry(logrus.New()))
	k.SetClusterID("test-cluster")

	if err := k.Ensure(); err != nil {
		k.Log.Fatalf("error ensuring kubernetes: %v", err)
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

// Init instance token for testing
func initInstanceToken(t *testing.T, vaultDev *vault_dev.VaultDev, dir string) *instanceToken.InstanceToken {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	log := logrus.NewEntry(logger)

	i := instanceToken.New(vaultDev.Client(), log)
	i.SetInitRole("")

	i.SetVaultConfigPath(dir)

	if _, err := os.Stat(i.InitTokenFilePath()); os.IsNotExist(err) {
		ifile, err := os.Create(i.InitTokenFilePath())
		if err != nil {
			t.Fatal(err)
		}
		defer ifile.Close()
	}

	_, err := os.Stat(i.TokenFilePath())
	if os.IsNotExist(err) {
		tfile, err := os.Create(i.TokenFilePath())
		if err != nil {
			t.Fatal(err)
		}
		defer tfile.Close()
	}

	i.WipeTokenFile(i.InitTokenFilePath())
	i.WipeTokenFile(i.TokenFilePath())

	return i
}
