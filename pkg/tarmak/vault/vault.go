// Copyright Jetstack Ltd. See LICENSE for details.
package vault

import (
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/google/uuid"
	vault "github.com/hashicorp/vault/api"
	vaultUnsealer "github.com/jetstack/vault-unsealer/pkg/vault"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

const (
	VaultStateSealed = iota
	VaultStateUnsealed
	VaultStateUnintialised
	VaultStateErr
	vaultTunnelCreationTimeoutSeconds = 100
)

const (
	Retries = 60
)

var _ interfaces.Vault = &Vault{}

type Vault struct {
	cluster interfaces.Cluster
	log     *logrus.Entry
}

func NewFromCluster(cluster interfaces.Cluster) (*Vault, error) {
	v := &Vault{
		cluster: cluster,
		log:     cluster.Log().WithField("module", "vault"),
	}
	return v, nil
}

// create tunnels
func (v *Vault) Tunnel() (interfaces.VaultTunnel, error) {
	return nil, nil // TODO: implement me
}

// path to the root token
func (v *Vault) rootTokenPath() string {
	return filepath.Join(v.cluster.Environment().ConfigPath(), "vault_root_token")
}

func (v *Vault) RootToken() (string, error) {
	path := v.rootTokenPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := utils.EnsureDirectory(filepath.Dir(path), 0700); err != nil {
			return "", fmt.Errorf("error creating directory: %s", err)
		}

		uuidValue := uuid.New()

		err := ioutil.WriteFile(path, []byte(fmt.Sprintf("%s\n", uuidValue.String())), 0600)
		if err != nil {
			return "", err
		}

		return uuidValue.String(), nil
	}

	uuidBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("unable to read vault root token %s: %s", path, err)
	}

	return strings.TrimSpace(string(uuidBytes)), nil
}

// returns the active vault tunnel for the whole cluster with provided FQDNs
func (v *Vault) TunnelFromFQDNs(vaultInternalFQDNs []string, vaultCA string) (interfaces.VaultTunnel, error) {

	tunnels, err := v.createTunnelsWithCA(vaultInternalFQDNs, vaultCA)
	if err != nil {
		return nil, err
	}

	activeNode := make(chan int, 0)

	var wg sync.WaitGroup
	for pos, _ := range tunnels {
		wg.Add(1)
		go func(pos int) {
			defer wg.Done()
			err := tunnels[pos].Start()
			if err != nil {
				v.log.Warn(err)
				return
			}
			health, err := tunnels[pos].VaultClient().Sys().Health()
			if err != nil {
				v.log.Warn(err)
				return
			}

			if health.Standby == false && health.Sealed == false && health.Initialized == true {
				activeNode <- pos
			}

		}(pos)
	}

	var activePos int
	select {
	case activePos = <-activeNode:
		v.log.Debug("active channel position recieved")
	case <-time.After(vaultTunnelCreationTimeoutSeconds * time.Second):
		return nil, fmt.Errorf("failed to retrieve active channel position")
	}

	go func(activePos int) {

		// wait for all tunnel attempts
		wg.Wait()

		// stop non-active tunnels
		for pos, _ := range tunnels {
			if pos == activePos {
				continue
			}
			tunnels[pos].Stop()
		}

	}(activePos)

	return tunnels[activePos], nil
}

func (v *Vault) createTunnelsWithCA(instances []string, vaultCA string) ([]*vaultTunnel, error) {
	certpool := x509.NewCertPool()
	ok := certpool.AppendCertsFromPEM([]byte(vaultCA))
	if !ok {
		return nil, fmt.Errorf("failed to parse vault CA. %q", vaultCA)
	}
	output := make([]*vaultTunnel, len(instances))
	for pos := range instances {
		fqdn := instances[pos]
		sshTunnel := v.cluster.Environment().Tarmak().SSH().Tunnel(
			fqdn, "8200", strconv.Itoa(utils.UnusedPort()), false,
		)
		vaultTunnel, err := NewTunnel(
			sshTunnel,
			fqdn,
			certpool,
		)
		if err != nil {
			return output, err
		}
		output[pos] = vaultTunnel
	}

	return output, nil
}

func (v *Vault) VerifyInitFromFQDNs(instances []string, vaultCA, vaultKMSKeyID, vaultUnsealKeyName string) error {

	rootToken, err := v.RootToken()
	if err != nil {
		return err
	}

	kv, err := v.cluster.Environment().Provider().VaultKVWithParams(vaultKMSKeyID, vaultUnsealKeyName)
	if err != nil {
		return err
	}

	tunnels, err := v.createTunnelsWithCA(instances, vaultCA)
	if err != nil {
		return err
	}

	var cl *vault.Client
	readyTunnelFunc := func() error {
		for pos, _ := range tunnels {
			err := tunnels[pos].Start()
			if err != nil {
				return err
			}
		}

		time.Sleep(time.Second * 2)

		for _, t := range tunnels {
			if t.Status() != VaultStateErr {
				cl = t.VaultClient()
				return nil
			}
		}

		for pos, _ := range tunnels {
			tunnels[pos].Stop()
		}

		return errors.New("failed to find a vault tunnel ready")
	}

	done := make(chan struct{})
	ctx := v.cluster.Environment().Tarmak().CancellationContext().TryOrCancel(done)
	constBackoff := backoff.NewConstantBackOff(time.Second * 5)
	b := backoff.WithContext(
		backoff.WithMaxTries(constBackoff, Retries),
		ctx)

	err = backoff.Retry(readyTunnelFunc, b)
	close(done)
	if err != nil {
		return fmt.Errorf("failed to obtain vault tunnel: %s", err)
	}

	initVaultFunc := func() error {
		health, err := cl.Sys().Health()
		if err != nil {
			err = fmt.Errorf("failed to get vault status: %s", err)
			v.log.Warn(err)
			return err
		}

		if !health.Sealed {
			return nil

		} else if !health.Initialized {
			unsealer, err := vaultUnsealer.New(kv, cl, vaultUnsealer.Config{
				KeyPrefix: "vault",

				SecretShares:    1,
				SecretThreshold: 1,

				InitRootToken:  rootToken,
				StoreRootToken: false,

				OverwriteExisting: true,
			})
			if err != nil {
				return fmt.Errorf("error creating new unsealer connection: %s", err)
			}

			if err := unsealer.Init(); err != nil {
				return fmt.Errorf("error initialising vault: %s", err)
			}

			v.log.Info("vault successfully initialised")

			return nil

		} else if health.Sealed {
			err := errors.New("a quorum of vault instances is sealed, retrying")
			v.log.Debug(err)
			return err
		} else {
			err := errors.New("a quorum of vault instances is in unknown state, retrying")
			v.log.Debug(err)
			return err
		}
	}

	done = make(chan struct{})
	ctx = v.cluster.Environment().Tarmak().CancellationContext().TryOrCancel(done)
	constBackoff = backoff.NewConstantBackOff(time.Second * 5)
	b = backoff.WithContext(
		backoff.WithMaxTries(constBackoff, Retries),
		ctx)

	err = backoff.Retry(initVaultFunc, b)
	close(done)
	if err != nil {
		return fmt.Errorf("time out verifying that vault cluster is initialised and unsealed: %s", err)
	}

	return nil
}
