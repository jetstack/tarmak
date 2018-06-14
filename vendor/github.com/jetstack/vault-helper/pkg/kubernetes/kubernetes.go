// Copyright Jetstack Ltd. See LICENSE for details.
package kubernetes

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/hashicorp/go-multierror"
	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

const FlagMaxValidityAdmin = "max-validity-admin"
const FlagMaxValidityCA = "max-validity-ca"
const FlagMaxValidityComponents = "max-validity-components"

const FlagInitTokenEtcd = "init-token-etcd"
const FlagInitTokenAll = "init-token-all"
const FlagInitTokenMaster = "init-token-master"
const FlagInitTokenWorker = "init-token-worker"

var Version string

type Backend interface {
	Ensure() error
	EnsureDryRun() (bool, error)
	Delete() error
	Path() string
	Type() string
	Name() string
}

type VaultLogical interface {
	Write(path string, data map[string]interface{}) (*vault.Secret, error)
	Read(path string) (*vault.Secret, error)
	Delete(path string) (*vault.Secret, error)
}

type VaultSys interface {
	ListMounts() (map[string]*vault.MountOutput, error)
	ListPolicies() ([]string, error)

	Mount(path string, mountInfo *vault.MountInput) error
	PutPolicy(name, rules string) error
	TuneMount(path string, config vault.MountConfigInput) error
	GetPolicy(name string) (string, error)

	Unmount(path string) error
	DeletePolicy(policy string) error
	Revoke(id string) error
}

type VaultAuth interface {
	Token() VaultToken
}

type VaultToken interface {
	CreateOrphan(opts *vault.TokenCreateRequest) (*vault.Secret, error)
	RevokeOrphan(token string) error
	Lookup(token string) (*vault.Secret, error)
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

type FlagInitTokens struct {
	Etcd   string
	Master string
	Worker string
	All    string
}

type Kubernetes struct {
	clusterID   string // clusterID is required parameter, lowercase only, [a-z0-9-]+
	vaultClient Vault
	Log         *logrus.Entry

	// PKI for kubernetes' state storage in Etcd
	etcdKubernetesBackend *PKIVaultBackend

	// PKI for the overlay network's state storage in Etcd
	etcdOverlayBackend *PKIVaultBackend

	// This is the core kubernetes PKI, which is used to authenticate all
	// kubernetes components.
	kubernetesBackend *PKIVaultBackend

	// This is a separate kubernetes PKI, it is used to authenticate request
	// headers proxied through the API server. This is utilized for API server
	// aggregation.
	kubernetesAPIProxyBackend *PKIVaultBackend

	// A generic vault backend for static secrets
	secretsBackend *GenericVaultBackend

	MaxValidityAdmin      time.Duration
	MaxValidityComponents time.Duration
	MaxValidityCA         time.Duration
	MaxValidityInitTokens time.Duration

	FlagInitTokens FlagInitTokens

	initTokens []*InitToken

	version string
}

var _ Backend = &PKIVaultBackend{}
var _ Backend = &GenericVaultBackend{}

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

	f = func(r rune) bool { return ((r < 'a' || r > 'z') && (r < '0' || r > '9')) && r != '-' }

	if strings.IndexFunc(clusterID, f) != -1 {
		return errors.New("Not a valid cluster ID name")
	}

	return nil
}

func New(vaultClient *vault.Client, logger *logrus.Entry) *Kubernetes {

	k := &Kubernetes{
		// set default validity periods
		MaxValidityCA:         time.Hour * 24 * 365 * 20, // Validity period of CA certificates
		MaxValidityComponents: time.Hour * 24 * 30,       // Validity period of Component certificates
		MaxValidityAdmin:      time.Hour * 24 * 365,      // Validity period of Admin ceritficate
		MaxValidityInitTokens: time.Hour * 24 * 365 * 5,  // Validity of init tokens
		FlagInitTokens: FlagInitTokens{
			Etcd:   "",
			Master: "",
			Worker: "",
			All:    "",
		},
		version: Version,
	}

	if vaultClient != nil {
		k.vaultClient = realVaultFromAPI(vaultClient)
	}
	if logger != nil {
		k.Log = logger
	}

	k.etcdKubernetesBackend = NewPKIVaultBackend(k, "etcd-k8s", k.Log)
	k.etcdOverlayBackend = NewPKIVaultBackend(k, "etcd-overlay", k.Log)
	k.kubernetesBackend = NewPKIVaultBackend(k, "k8s", k.Log)
	k.kubernetesAPIProxyBackend = NewPKIVaultBackend(k, "k8s-api-proxy", k.Log)

	k.secretsBackend = k.NewGenericVaultBackend(k.Log)

	return k
}

func (k *Kubernetes) SetClusterID(clusterID string) {
	k.clusterID = clusterID
}

func (k *Kubernetes) backends() []Backend {
	return []Backend{
		k.etcdKubernetesBackend,
		k.etcdOverlayBackend,
		k.kubernetesBackend,
		k.kubernetesAPIProxyBackend,
		k.secretsBackend,
	}
}

func (k *Kubernetes) Ensure() error {
	if err := isValidClusterID(k.clusterID); err != nil {
		return fmt.Errorf("error '%s' is not a valid clusterID", k.clusterID)
	}

	// setup backends
	var result *multierror.Error
	for _, backend := range k.backends() {
		if err := backend.Ensure(); err != nil {
			result = multierror.Append(result, fmt.Errorf("backend %s: %s", backend.Path(), err))
		}
	}
	if result != nil {
		return result.ErrorOrNil()
	}

	// setup pki roles
	if err := k.ensurePKIRolesEtcd(k.etcdKubernetesBackend); err != nil {
		result = multierror.Append(result, err)
	}
	if err := k.ensurePKIRolesEtcd(k.etcdOverlayBackend); err != nil {
		result = multierror.Append(result, err)
	}
	if err := k.ensurePKIRolesK8S(k.kubernetesBackend); err != nil {
		result = multierror.Append(result, err)
	}
	if err := k.ensurePKIRolesK8SAPIProxy(k.kubernetesAPIProxyBackend); err != nil {
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

	return result.ErrorOrNil()
}

type DryRun struct {
	*multierror.Error
}

func (d *DryRun) changeNeeded(change bool, err error) bool {
	if err != nil {
		d.Error = multierror.Append(d.Error, err)
	}

	return change
}

// return true if change needed
func (k *Kubernetes) EnsureDryRun() (bool, error) {
	d := &DryRun{
		new(multierror.Error),
	}

	if len(k.initTokens) == 0 {
		k.initTokens = k.NewInitTokens()
	}

	for _, b := range k.backends() {
		if d.changeNeeded(b.EnsureDryRun()) {
			return true, d.ErrorOrNil()
		}
	}

	if d.changeNeeded(k.ensureDryRunPKIRolesEtcd(k.etcdKubernetesBackend)) {
		return true, d.ErrorOrNil()
	}

	if d.changeNeeded(k.ensureDryRunPKIRolesEtcd(k.etcdOverlayBackend)) {
		return true, d.ErrorOrNil()
	}

	if d.changeNeeded(k.ensureDryRunPKIRolesK8S(k.kubernetesBackend)) {
		return true, d.ErrorOrNil()
	}

	if d.changeNeeded(k.ensureDryRunPKIRolesK8SAPIProxy(k.kubernetesAPIProxyBackend)) {
		return true, d.ErrorOrNil()
	}

	if d.changeNeeded(k.ensureDryRunPolicies()) {
		return true, d.ErrorOrNil()
	}

	if d.changeNeeded(k.ensureDryRunInitTokens()) {
		return true, d.ErrorOrNil()
	}

	return false, d.ErrorOrNil()
}

func (k *Kubernetes) Delete() error {
	var result *multierror.Error

	if err := k.deletePolicies(); err != nil {
		result = multierror.Append(result, err)
	}

	for _, i := range k.initTokens {
		if err := i.Delete(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	if err := k.deletePKIRolesEtcd(k.etcdKubernetesBackend); err != nil {
		result = multierror.Append(result, err)
	}
	if err := k.deletePKIRolesEtcd(k.etcdOverlayBackend); err != nil {
		result = multierror.Append(result, err)
	}
	if err := k.deletePKIRolesK8S(k.kubernetesBackend); err != nil {
		result = multierror.Append(result, err)
	}
	if err := k.deletePKIRolesK8SAPIProxy(k.kubernetesAPIProxyBackend); err != nil {
		result = multierror.Append(result, err)
	}

	for _, b := range k.backends() {
		if err := b.Delete(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func (k *Kubernetes) Path() string {
	return k.clusterID
}

func (k *Kubernetes) NewGenericVaultBackend(logger *logrus.Entry) *GenericVaultBackend {
	return &GenericVaultBackend{
		kubernetes: k,
		initTokens: make(map[string]string),
		Log:        logger,
	}
}

func GetMountByPath(vaultClient Vault, mountPath string) (*vault.MountOutput, error) {
	mountPath = filepath.Clean(mountPath)

	mounts, err := vaultClient.Sys().ListMounts()
	if err != nil {
		return nil, fmt.Errorf("error listing mounts: %v", err)
	}

	for key, _ := range mounts {
		if filepath.Clean(key) == mountPath || filepath.Dir(key) == mountPath {
			return mounts[key], nil
		}
	}

	return nil, nil
}

func (k *Kubernetes) NewInitToken(role, expected string, policies []string) *InitToken {
	return &InitToken{
		Role:          role,
		Policies:      policies,
		kubernetes:    k,
		ExpectedToken: expected,
	}
}

func (k *Kubernetes) NewInitTokens() []*InitToken {
	var initTokens []*InitToken

	initTokens = append(initTokens, k.NewInitToken("etcd", k.FlagInitTokens.Etcd, []string{
		k.etcdPolicy().Name,
	}))
	initTokens = append(initTokens, k.NewInitToken("master", k.FlagInitTokens.Master, []string{
		k.masterPolicy().Name,
		k.workerPolicy().Name,
	}))
	initTokens = append(initTokens, k.NewInitToken("worker", k.FlagInitTokens.Worker, []string{
		k.workerPolicy().Name,
	}))
	initTokens = append(initTokens, k.NewInitToken("all", k.FlagInitTokens.All, []string{
		k.etcdPolicy().Name,
		k.masterPolicy().Name,
		k.workerPolicy().Name,
	}))

	return initTokens
}

func (k *Kubernetes) ensureInitTokens() error {
	var result *multierror.Error

	k.initTokens = append(k.initTokens, k.NewInitTokens()...)

	for _, initToken := range k.initTokens {
		if err := initToken.Ensure(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (k *Kubernetes) ensureDryRunInitTokens() (bool, error) {
	if len(k.initTokens) == 0 {
		return true, nil
	}

	var result *multierror.Error
	for _, i := range k.initTokens {
		changeNeeded, err := i.EnsureDryRun()
		if err != nil {
			result = multierror.Append(result, err)
		}
		if changeNeeded {
			return true, result.ErrorOrNil()
		}
	}

	return false, result.ErrorOrNil()
}

func (k *Kubernetes) InitTokens() map[string]string {
	output := map[string]string{}
	for _, initToken := range k.initTokens {
		token, err := initToken.InitToken()
		if err == nil {
			output[initToken.Role] = token
		}
	}
	return output
}

func (k *Kubernetes) SetInitFlags(flags FlagInitTokens) {
	k.FlagInitTokens = flags
}

func (k *Kubernetes) Version() string {
	return k.version
}

func (k *Kubernetes) SetVersion(version string) {
	k.version = version
}
