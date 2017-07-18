package stack

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"sync"

	vault "github.com/hashicorp/vault/api"
	vaultUnsealer "github.com/jetstack-experimental/vault-unsealer/pkg/vault"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

var vaultClientLock sync.Mutex

type VaultStack struct {
	*Stack
}

var _ interfaces.Stack = &VaultStack{}

func newVaultStack(s *Stack, conf *config.StackVault) (*VaultStack, error) {
	s.name = config.StackNameVault
	return &VaultStack{
		Stack: s,
	}, nil
}

func (s *VaultStack) Variables() map[string]interface{} {
	return map[string]interface{}{}
}

const (
	VaultStateSealed = iota
	VaultStateUnsealed
	VaultStateUnintialised
	VaultStateErr
)

type vaultTunnel struct {
	tunnel      interfaces.Tunnel
	tunnelError error
	client      *vault.Client
	fqdn        string
}

func (s *VaultStack) vaultCA() ([]byte, error) {
	vaultCAIntf, ok := s.output["vault_ca"]
	if !ok {
		return []byte{}, fmt.Errorf("unable to find terraform output 'vault_ca'")
	}

	vaultCA, ok := vaultCAIntf.(string)
	if !ok {
		return []byte{}, fmt.Errorf("unexpected type for 'vault_ca': %t", vaultCAIntf)
	}

	return []byte(vaultCA), nil
}

func (s *VaultStack) vaultInstanceFQDNs() ([]string, error) {
	instanceFQDNsIntf, ok := s.output["instance_fqdns"]
	if !ok {
		return []string{}, fmt.Errorf("unable to find terraform output 'instance_fqdns'")
	}

	instanceFQDNsInftSlice, ok := instanceFQDNsIntf.([]interface{})
	if !ok {
		return []string{}, fmt.Errorf("unexpected type for 'instance_fqdns': %T", instanceFQDNsIntf)
	}

	instanceFQDNs := make([]string, len(instanceFQDNsInftSlice))
	for pos, value := range instanceFQDNsInftSlice {
		var ok bool
		instanceFQDNs[pos], ok = value.(string)
		if !ok {
			return []string{}, fmt.Errorf("unexpected type for element %d in 'instance_fqdns': %T", pos, value)
		}
	}

	return instanceFQDNs, nil
}

func (s *VaultStack) vaultTunnels() ([]*vaultTunnel, error) {
	vaultCA, err := s.vaultCA()
	if err != nil {
		return []*vaultTunnel{}, fmt.Errorf("couldn't load vault CA from terraform: %s", err)
	}

	tlsConfig := &tls.Config{RootCAs: x509.NewCertPool()}

	ok := tlsConfig.RootCAs.AppendCertsFromPEM(vaultCA)
	if !ok {
		return []*vaultTunnel{}, fmt.Errorf("couldn't load vault CA certificate into http client")
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := &http.Client{Transport: tr}

	vaultClient, err := vault.NewClient(&vault.Config{
		HttpClient: httpClient,
	})
	if err != nil {
		return []*vaultTunnel{}, fmt.Errorf("couldn't init vault client: %s:", err)
	}

	vaultInstances, err := s.vaultInstanceFQDNs()
	if err != nil {
		return []*vaultTunnel{}, fmt.Errorf("couldn't load vault instance fqdns from terraform: %s", err)
	}

	output := make([]*vaultTunnel, len(vaultInstances))
	for pos, _ := range vaultInstances {
		output[pos] = s.newVaultTunnel(vaultInstances[pos], vaultClient)
	}

	return output, nil

}

func (s *VaultStack) newVaultTunnel(fqdn string, client *vault.Client) *vaultTunnel {
	return &vaultTunnel{
		tunnel: s.Context().Environment().Tarmak().SSH().Tunnel("bastion", fqdn, 8200),
		client: client,
		fqdn:   fqdn,
	}
}

func (v *vaultTunnel) FQDN() string {
	return v.fqdn
}
func (v *vaultTunnel) Start() error {

	if err := v.tunnel.Start(); err != nil {
		v.tunnelError = err
		return err
	}

	return nil
}

func (v *vaultTunnel) Stop() error {
	return v.tunnel.Stop()
}

func (v *vaultTunnel) VaultClient() *vault.Client {
	v.client.SetAddress(fmt.Sprintf("https://localhost:%d", v.tunnel.Port()))
	return v.client
}

func (v *vaultTunnel) Status() int {
	if v.tunnelError != nil {
		return VaultStateErr
	}

	vaultClientLock.Lock()
	defer vaultClientLock.Unlock()

	v.VaultClient()

	initStatus, err := v.client.Sys().InitStatus()
	if err != nil {
		return VaultStateErr
	}

	if !initStatus {
		return VaultStateUnintialised
	}

	sealStatus, err := v.client.Sys().SealStatus()
	if err != nil {
		return VaultStateErr
	}

	if sealStatus.Sealed {
		return VaultStateSealed
	}
	return VaultStateUnsealed
}

func (s *VaultStack) VerifyPost() error {

	tunnels, err := s.vaultTunnels()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for pos, _ := range tunnels {
		wg.Add(1)
		go func(pos int) {
			defer wg.Done()
			err := tunnels[pos].Start()
			if err != nil {
				s.log.Warn(err)
			}
		}(pos)
	}

	// wait for all tunnel attempts
	wg.Wait()

	// get state of all instances
	instanceState := map[int][]*vaultTunnel{}
	for pos, _ := range tunnels {
		state := tunnels[pos].Status()
		if _, ok := instanceState[state]; !ok {
			instanceState[state] = []*vaultTunnel{tunnels[pos]}
		} else {
			instanceState[state] = append(instanceState[state], tunnels[pos])
		}
		s.log.Debugf("vault %s status: %d", tunnels[pos].FQDN(), tunnels[pos].Status())
	}

	// get state that has quorum
	for state, instances := range instanceState {
		if len(instances) > len(tunnels)/2 {
			if state == VaultStateUnsealed {
				return nil
			} else if state == VaultStateUnintialised {
				kv, err := s.Context().Environment().Provider().VaultKV()
				if err != nil {
					return err
				}

				vaultClientLock.Lock()
				defer vaultClientLock.Unlock()

				cl := instances[0].VaultClient()

				v, err := vaultUnsealer.New(kv, cl, vaultUnsealer.Config{
					KeyPrefix: "vault",

					SecretShares:    1,
					SecretThreshold: 1,

					// TODO: use random UUID here
					InitRootToken:  "root-token",
					StoreRootToken: false,

					OverwriteExisting: true,
				})

				err = v.Init()
				if err != nil {
					return fmt.Errorf("error initialising vault: %s", err)
				}

			} else if state == VaultStateSealed {
				return fmt.Errorf("a quorum of vault instances is sealed")
			}
			return fmt.Errorf("a quorum of vault instances is in unknown state")
		}
	}

	defer func() {
		var wg sync.WaitGroup
		for pos, _ := range tunnels {
			wg.Add(1)
			go func(pos int) {
				defer wg.Done()
				err := tunnels[pos].Stop()
				if err != nil {
					s.log.Warn(err)
				}
			}(pos)
		}
		wg.Wait()
	}()

	return nil
}
