package cert

import (
	"encoding/pem"
	"fmt"
	"os"
	"os/user"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/jetstack/vault-helper/pkg/instanceToken"
)

const FlagKeyBitSize = "key-bit-size"
const FlagKeyType = "key-type"
const FlagIpSans = "ip-sans"
const FlagSanHosts = "san-hosts"
const FlagOwner = "owner"
const FlagGroup = "group"
const FlagOrganisation = "organisation"

type Cert struct {
	role         string
	commonName   string
	organisation []string
	destination  string
	bitSize      int
	pemSize      int
	keyType      string
	ipSans       []string
	sanHosts     []string
	owner        string
	group        string
	data         *pem.Block

	Log           *logrus.Entry
	instanceToken *instanceToken.InstanceToken
}

func (c *Cert) RunCert() error {
	if err := c.EnsureKey(); err != nil {
		return fmt.Errorf("error ensuring key: %v", err)
	}

	//if err := c.TokenRenew(); err != nil {
	//	return fmt.Errorf("error renewing tokens: %v", err)
	//}

	if err := c.RequestCertificate(); err != nil {
		return fmt.Errorf("error requesting certificate: %v", err)
	}

	return nil
}

//func (c *Cert) TokenRenew() error {
//	i := instanceToken.New(c.vaultClient, c.Log)
//
//	return i.TokenRenewRun()
//}

func (c *Cert) DeleteFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("error removing file at '%s':  %v", path, err)
	}

	return nil
}

func (c *Cert) WritePermissions(path string, perm os.FileMode) error {
	if err := os.Chmod(path, perm); err != nil {
		return fmt.Errorf("failed to change permissons of file '%s' to 0600: %v", path, err)
	}

	var uid int
	var gid int
	var err error
	var curr *user.User

	if c.Owner() == "" {
		c.Log.Debugf("No owner given. Defaulting permissions to current user")
		if curr, err = user.Current(); err != nil {
			return fmt.Errorf("error retrieving current user info: %v", err)
		}

		if uid, err = strconv.Atoi(curr.Uid); err != nil {
			return fmt.Errorf("failed to convert user uid '%s' (string) to (int): %v", curr.Uid, err)
		}

	} else if u, err := strconv.Atoi(c.Owner()); err == nil {
		c.Log.Debugf("User is a number. Using instead of lookup user")
		uid = u

	} else {
		usr, err := user.Lookup(c.Owner())
		if err != nil {
			return fmt.Errorf("failed to find user '%s' on system: %v", c.Owner(), err)
		}

		if uid, err = strconv.Atoi(usr.Uid); err != nil {
			return fmt.Errorf("failed to convert user uid '%s' (string) to (int): %v", usr.Uid, err)
		}
	}

	if c.Group() == "" {
		c.Log.Debugf("No group given. Defaulting permissions to current user-group")
		if curr == nil {
			if curr, err = user.Current(); err != nil {
				return fmt.Errorf("error retrieving current user info: %v", err)
			}
		}

		if gid, err = strconv.Atoi(curr.Gid); err != nil {
			return fmt.Errorf("failed to convert user gid '%s' (string) to (int): %v", curr.Gid, err)
		}

	} else if g, err := strconv.Atoi(c.Group()); err == nil {
		c.Log.Debugf("Group is a number. Using as gid instead of lookup group")
		gid = g

	} else {
		grp, err := user.LookupGroup(c.Group())
		if err != nil {
			return fmt.Errorf("failed to find group '%s' on system: %v", c.Group(), err)
		}

		if gid, err = strconv.Atoi(grp.Gid); err != nil {
			return fmt.Errorf("failed to convert group gid '%s' (string) to (int): %v", grp.Gid, err)
		}
	}

	if err := os.Chown(path, uid, gid); err != nil {
		return fmt.Errorf("failed to change group and owner of file '%s' to usr:'%s' grp:'%s': %v", path, c.Owner(), c.Group(), err)
	}

	c.Log.Debugf("Set permissons on file: %s", path)

	return nil
}

func New(logger *logrus.Entry, i *instanceToken.InstanceToken) *Cert {
	c := &Cert{
		bitSize:       2048,
		keyType:       "RSA",
		instanceToken: i,
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

func (c *Cert) SetOrganisation(org []string) {
	c.organisation = org
}

func (c *Cert) Organisation() []string {
	return c.organisation
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

func (c *Cert) SetInstanceToken(i *instanceToken.InstanceToken) {
	c.instanceToken = i
}
func (c *Cert) InstanceToken() *instanceToken.InstanceToken {
	return c.instanceToken
}
