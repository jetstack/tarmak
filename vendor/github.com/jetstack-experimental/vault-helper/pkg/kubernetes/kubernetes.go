package kubernetes

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"
	vault "github.com/hashicorp/vault/api"
)

type Backend interface {
	Ensure() (bool, error)
	Path() string
}

type VaultLogical interface {
	Write(path string, data map[string]interface{}) (*vault.Secret, error)
	Read(path string) (*vault.Secret, error)
}

type VaultSys interface {
	ListMounts() (map[string]*vault.MountOutput, error)
	ListPolicies() ([]string, error)

	Mount(path string, mountInfo *vault.MountInput) error
	PutPolicy(name, rules string) error
	TuneMount(path string, config vault.MountConfigInput) error
	GetPolicy(name string) (string, error)
}

type VaultAuth interface {
	Token() VaultToken
}

type VaultToken interface {
	CreateOrphan(opts *vault.TokenCreateRequest) (*vault.Secret, error)
}

type Vault interface {
	Logical() VaultLogical
	Sys() VaultSys
	Auth() VaultAuth
}

type realVault struct {
	c *vault.Client
}

type realVaultAuth struct {
	a *vault.Auth
}

func (rv *realVault) Auth() VaultAuth {
	return &realVaultAuth{a: rv.c.Auth()}
}
func (rv *realVault) Sys() VaultSys {
	return rv.c.Sys()
}
func (rv *realVault) Logical() VaultLogical {
	return rv.c.Logical()
}

func (rva *realVaultAuth) Token() VaultToken {
	return rva.a.Token()
}

func realVaultFromAPI(vaultClient *vault.Client) Vault {
	return &realVault{c: vaultClient}
}

type Kubernetes struct {
	clusterID   string // clusterID is required parameter, lowercase only, [a-z0-9-]+
	vaultClient Vault

	etcdKubernetesPKI *PKI
	etcdOverlayPKI    *PKI
	kubernetesPKI     *PKI
	secretsGeneric    *Generic

	MaxValidityAdmin      time.Duration
	MaxValidityComponents time.Duration
	MaxValidityCA         time.Duration
	MaxValidityInitTokens time.Duration

	initTokens []*InitToken
}

var _ Backend = &PKI{}
var _ Backend = &Generic{}

func isValidClusterID(clusterID string) error {

	if len(clusterID) < 1 {
		return errors.New("Invalid cluster ID - None given")
	}

	if !unicode.IsLetter([]rune(clusterID)[0]) {
		return errors.New("First character is not a valid character")
	}

	f := func(r rune) bool {
		return ((r < 'a' || r > 'z') && (r < '0' || r > '9') && (r >= 'A' || r <= 'Z')) && r != '-'
	}

	if strings.IndexFunc(clusterID, f) != -1 {
		return errors.New("Invalid cluster ID - contains uppercase")
	}

	f = func(r rune) bool {
		return ((r < 'a' || r > 'z') && (r < '0' || r > '9')) && r != '-'
	}

	if strings.IndexFunc(clusterID, f) != -1 {
		return errors.New("Not a valid cluster ID name")
	}

	return nil

}

func New(vaultClient *vault.Client) *Kubernetes {

	k := &Kubernetes{
		// set default validity periods
		MaxValidityCA:         time.Hour * 24 * 365 * 20, // Validity period of CA certificates
		MaxValidityComponents: time.Hour * 24 * 30,       // Validity period of Component certificates
		MaxValidityAdmin:      time.Hour * 24 * 365,      // Validity period of Admin ceritficate
		MaxValidityInitTokens: time.Hour * 24 * 365 * 5,  // Validity of init tokens
	}

	if vaultClient != nil {
		k.vaultClient = realVaultFromAPI(vaultClient)
	}

	k.etcdKubernetesPKI = NewPKI(k, "etcd-k8s")
	k.etcdOverlayPKI = NewPKI(k, "etcd-overlay")
	k.kubernetesPKI = NewPKI(k, "k8s")

	k.secretsGeneric = k.NewGeneric()

	return k
}

func (k *Kubernetes) SetClusterID(clusterID string) {
	k.clusterID = clusterID
}

func (k *Kubernetes) backends() []Backend {
	return []Backend{
		k.etcdKubernetesPKI,
		k.etcdOverlayPKI,
		k.kubernetesPKI,
		k.secretsGeneric,
	}
}

func (k *Kubernetes) Ensure() error {
	if err := isValidClusterID(k.clusterID); err != nil {
		return fmt.Errorf("error '%s' is not a valid clusterID", k.clusterID)
	}

	// setup backends
	var result error
	change := false
	str := "Mounted & CA written for: "
	for _, backend := range k.backends() {
		if changed, err := backend.Ensure(); err != nil {
			result = multierror.Append(result, fmt.Errorf("backend %s: %s", backend.Path(), err))
		} else {
			if changed {
				change = true
				str += "'" + backend.Path() + "'  "
			}
		}
	}
	if change {
		logrus.Infof(str)
	}
	if result != nil {
		return result
	}

	// setup pki roles
	if err := k.ensurePKIRolesEtcd(k.etcdKubernetesPKI); err != nil {
		result = multierror.Append(result, err)
	}
	if err := k.ensurePKIRolesEtcd(k.etcdOverlayPKI); err != nil {
		result = multierror.Append(result, err)
	}
	if err := k.ensurePKIRolesK8S(k.kubernetesPKI); err != nil {
		result = multierror.Append(result, err)
	}

	// setup policies
	if err := k.ensurePolicies(); err != nil {
		result = multierror.Append(result, err)
	}

	// setup init tokens
	if err := k.ensureInitTokens(); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (k *Kubernetes) Path() string {
	return k.clusterID
}

func (k *Kubernetes) NewGeneric() *Generic {
	return &Generic{
		kubernetes: k,
		initTokens: make(map[string]string),
	}
}

func GetMountByPath(vaultClient Vault, mountPath string) (*vault.MountOutput, error) {

	mounts, err := vaultClient.Sys().ListMounts()
	if err != nil {
		return nil, fmt.Errorf("error listing mounts: %s", err)
	}

	var mount *vault.MountOutput
	for key, _ := range mounts {
		if filepath.Clean(key) == filepath.Clean(mountPath) {
			mount = mounts[key]
			break
		}
	}

	return mount, nil
}

func (k *Kubernetes) NewInitToken(role string, policies []string) *InitToken {
	return &InitToken{
		Role:       role,
		Policies:   policies,
		kubernetes: k,
	}
}

func (k *Kubernetes) ensureInitTokens() error {
	var result error

	k.initTokens = append(k.initTokens, k.NewInitToken("etcd", []string{
		k.etcdPolicy().Name,
	}))
	k.initTokens = append(k.initTokens, k.NewInitToken("master", []string{
		k.masterPolicy().Name,
		k.workerPolicy().Name,
	}))
	k.initTokens = append(k.initTokens, k.NewInitToken("worker", []string{
		k.workerPolicy().Name,
	}))
	k.initTokens = append(k.initTokens, k.NewInitToken("all", []string{
		k.etcdPolicy().Name,
		k.masterPolicy().Name,
		k.workerPolicy().Name,
	}))

	change := false
	strc := "Init_tokens created for: "
	strw := "Init_tokens written for: "
	for _, initToken := range k.initTokens {
		if changed, err := initToken.Ensure(); err != nil {
			result = multierror.Append(result, err)
		} else {
			strw += "'" + initToken.Name() + "'  "
			if changed {
				change = true
				strc += "'" + initToken.Name() + "'  "
			}
		}
	}
	if change {
		logrus.Infof(strc)
	}
	logrus.Infof(strw)

	return result
}

func (k *Kubernetes) InitTokens() map[string]string {
	output := map[string]string{}
	for _, initToken := range k.initTokens {
		token, _, err := initToken.InitToken()
		if err == nil {
			output[initToken.Role] = token
		}
	}
	return output
}
