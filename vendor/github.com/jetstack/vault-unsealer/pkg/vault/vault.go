package vault

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"

	"github.com/jetstack/vault-unsealer/pkg/kv"
)

// That configures the vault API

type Config struct {
	// key prefix
	KeyPrefix string

	// how many key parts exist
	SecretShares int
	// how many of these parts are needed to unseal vault  (secretThreshold <= secretShares)
	SecretThreshold int

	// if this root token is set, the dynamic generated will be invalidated and this created instead
	InitRootToken string
	// should the root token be stored in the keyStore
	StoreRootToken bool

	// overwrite existing tokens
	OverwriteExisting bool
}

// vault is an implementation of the Vault interface that will perform actions
// against a Vault server, using a provided KMS to retreive
type vault struct {
	keyStore kv.Service
	cl       *api.Client
	config   *Config
}

var _ Vault = &vault{}

// Vault is an interface that can be used to attempt to perform actions against
// a Vault server.
type Vault interface {
	Sealed() (bool, error)
	Unseal() error
	Init() error
}

// New returns a new vault Vault, or an error.
func New(k kv.Service, cl *api.Client, config Config) (Vault, error) {

	if config.SecretShares < config.SecretThreshold {
		return nil, errors.New("the secret threshold can't be bigger than the shares")
	}

	return &vault{
		keyStore: k,
		cl:       cl,
		config:   &config,
	}, nil
}

func (u *vault) Sealed() (bool, error) {
	resp, err := u.cl.Sys().SealStatus()
	if err != nil {
		return false, fmt.Errorf("error checking status: %s", err.Error())
	}
	return resp.Sealed, nil
}

// Unseal will attempt to unseal vault by retrieving keys from the kms service
// and sending unseal requests to vault. It will return an error if retrieving
// a key fails, or if the unseal progress is reset to 0 (indicating that a key)
// was invalid.
func (u *vault) Unseal() error {
	for i := 0; ; i++ {
		keyID := u.unsealKeyForID(i)

		logrus.Debugf("retrieving key from kms service...")
		k, err := u.keyStore.Get(keyID)

		if err != nil {
			return fmt.Errorf("unable to get key '%s': %s", keyID, err.Error())
		}

		logrus.Debugf("sending unseal request to vault...")
		resp, err := u.cl.Sys().Unseal(string(k))

		if err != nil {
			return fmt.Errorf("fail to send unseal request to vault: %s", err.Error())
		}

		logrus.Debugf("got unseal response: %+v", *resp)

		if !resp.Sealed {
			return nil
		}

		// if progress is 0, we failed to unseal vault.
		if resp.Progress == 0 {
			return fmt.Errorf("failed to unseal vault. progress reset to 0")
		}
	}
}

func (u *vault) keyStoreNotFound(key string) bool {
	_, err := u.keyStore.Get(key)
	if _, ok := err.(*kv.NotFoundError); ok {
		return true
	}
	return false
}

func (u *vault) keyStoreSet(key string, val []byte) error {
	if !u.config.OverwriteExisting && !u.keyStoreNotFound(key) {
		return fmt.Errorf("error setting key '%s': it already exists", key)
	}
	return u.keyStore.Set(key, val)
}

func (u *vault) Init() error {
	// test backend first
	err := u.keyStore.Test(u.testKey())
	if err != nil {
		return fmt.Errorf("error testing keystore before init: %s", err.Error())
	}

	// test for an existing keys
	if !u.config.OverwriteExisting {
		keys := []string{
			u.rootTokenKey(),
		}

		// add unseal keys
		for i := 0; i <= u.config.SecretShares; i++ {
			keys = append(keys, u.unsealKeyForID(i))
		}

		// test every key
		for _, key := range keys {
			if !u.keyStoreNotFound(key) {
				return fmt.Errorf("error before init: keystore value for '%s' already exists", key)
			}
		}
	}

	resp, err := u.cl.Sys().Init(&api.InitRequest{
		SecretShares:    u.config.SecretShares,
		SecretThreshold: u.config.SecretThreshold,
	})

	if err != nil {
		return fmt.Errorf("error initialising vault: %s", err.Error())
	}

	for i, k := range resp.Keys {
		keyID := u.unsealKeyForID(i)
		err := u.keyStoreSet(keyID, []byte(k))

		if err != nil {
			return fmt.Errorf("error storing unseal key '%s': %s", keyID, err.Error())
		}
	}

	rootToken := resp.RootToken

	// this sets up a predefined root token
	if u.config.InitRootToken != "" {
		logrus.Info("setting up init root token, waiting for vault to be unsealed")

		count := 0
		wait := time.Second * 2
		for {
			sealed, err := u.Sealed()
			if !sealed {
				break
			}
			if err == nil {
				logrus.Info("vault still sealed, wait for unsealing")
			} else {
				logrus.Infof("vault not reachable: %s", err.Error())
			}

			count++
			time.Sleep(wait)
		}

		// use temporary token
		u.cl.SetToken(resp.RootToken)

		// setup root token with provided key
		_, err := u.cl.Auth().Token().CreateOrphan(&api.TokenCreateRequest{
			ID:          u.config.InitRootToken,
			Policies:    []string{"root"},
			DisplayName: "root-token",
			NoParent:    true,
		})
		if err != nil {
			return fmt.Errorf("unable to setup requested root token, (temporary root token: '%s'): %s", resp.RootToken, err)
		}

		// revoke the temporary token
		err = u.cl.Auth().Token().RevokeSelf(resp.RootToken)
		if err != nil {
			return fmt.Errorf("unable to revoke temporary root token: %s", err.Error())
		}

		rootToken = u.config.InitRootToken
	}

	if u.config.StoreRootToken {
		rootTokenKey := u.rootTokenKey()
		if err = u.keyStoreSet(rootTokenKey, []byte(resp.RootToken)); err != nil {
			return fmt.Errorf("error storing root token '%s' in key'%s'", rootToken, rootTokenKey)
		}
		logrus.WithField("key", rootTokenKey).Info("root token stored in key store")
	} else if u.config.InitRootToken == "" {
		logrus.WithField("root-token", resp.RootToken).Warnf("won't store root token in key store, this token grants full privileges to vault, so keep this secret")
	}

	return nil

}

func (u *vault) unsealKeyForID(i int) string {
	return fmt.Sprintf("%s-unseal-%d", u.config.KeyPrefix, i)
}

func (u *vault) rootTokenKey() string {
	return fmt.Sprintf("%s-root", u.config.KeyPrefix)
}

func (u *vault) testKey() string {
	return fmt.Sprintf("%s-test", u.config.KeyPrefix)
}
