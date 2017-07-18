package environment

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/golang/mock/gomock"
	"golang.org/x/crypto/ssh"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/mocks"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var fakeSSHKeyInsecurePrivate = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEA6NF8iallvQVp22WDkTkyrtvp9eWW6A8YVr+kz4TjGYe7gHzI
w+niNltGEFHzD8+v1I2YJ6oXevct1YeS0o9HZyN1Q9qgCgzUFtdOKLv6IedplqoP
kcmF0aYet2PkEDo3MlTBckFXPITAMzF8dJSIFo9D8HfdOV0IAdx4O7PtixWKn5y2
hMNG0zQPyUecp4pzC6kivAIhyfHilFR61RGL+GPXQ2MWZWFYbAGjyiYJnAmCP3NO
Td0jMZEnDkbUvxhMmBYSdETk1rRgm+R4LOzFUGaHqHDLKLX+FIPKcF96hrucXzcW
yLbIbEgE98OHlnVYCzRdK8jlqm8tehUc9c9WhQIBIwKCAQEA4iqWPJXtzZA68mKd
ELs4jJsdyky+ewdZeNds5tjcnHU5zUYE25K+ffJED9qUWICcLZDc81TGWjHyAqD1
Bw7XpgUwFgeUJwUlzQurAv+/ySnxiwuaGJfhFM1CaQHzfXphgVml+fZUvnJUTvzf
TK2Lg6EdbUE9TarUlBf/xPfuEhMSlIE5keb/Zz3/LUlRg8yDqz5w+QWVJ4utnKnK
iqwZN0mwpwU7YSyJhlT4YV1F3n4YjLswM5wJs2oqm0jssQu/BT0tyEXNDYBLEF4A
sClaWuSJ2kjq7KhrrYXzagqhnSei9ODYFShJu8UWVec3Ihb5ZXlzO6vdNQ1J9Xsf
4m+2ywKBgQD6qFxx/Rv9CNN96l/4rb14HKirC2o/orApiHmHDsURs5rUKDx0f9iP
cXN7S1uePXuJRK/5hsubaOCx3Owd2u9gD6Oq0CsMkE4CUSiJcYrMANtx54cGH7Rk
EjFZxK8xAv1ldELEyxrFqkbE4BKd8QOt414qjvTGyAK+OLD3M2QdCQKBgQDtx8pN
CAxR7yhHbIWT1AH66+XWN8bXq7l3RO/ukeaci98JfkbkxURZhtxV/HHuvUhnPLdX
3TwygPBYZFNo4pzVEhzWoTtnEtrFueKxyc3+LjZpuo+mBlQ6ORtfgkr9gBVphXZG
YEzkCD3lVdl8L4cw9BVpKrJCs1c5taGjDgdInQKBgHm/fVvv96bJxc9x1tffXAcj
3OVdUN0UgXNCSaf/3A/phbeBQe9xS+3mpc4r6qvx+iy69mNBeNZ0xOitIjpjBo2+
dBEjSBwLk5q5tJqHmy/jKMJL4n9ROlx93XS+njxgibTvU6Fp9w+NOFD/HvxB3Tcz
6+jJF85D5BNAG3DBMKBjAoGBAOAxZvgsKN+JuENXsST7F89Tck2iTcQIT8g5rwWC
P9Vt74yboe2kDT531w8+egz7nAmRBKNM751U/95P9t88EDacDI/Z2OwnuFQHCPDF
llYOUI+SpLJ6/vURRbHSnnn8a/XG+nzedGH5JGqEJNQsz+xT2axM0/W/CRknmGaJ
kda/AoGANWrLCz708y7VYgAtW2Uf1DPOIYMdvo6fxIB5i9ZfISgcJ/bbCUkFrhoH
+vq/5CIWxCPp0f85R4qxxQ5ihxJ0YDQT9Jpx4TMss4PSavPaBH3RXow5Ohe+bYoQ
NE5OgEXk2wVfZczCZpigBKbKZHNYcelXtTt/nP3rsCuGcM4h53s=
-----END RSA PRIVATE KEY-----
`

func testDefaultEnvironmentConfig() *config.Environment {
	return &config.Environment{
		AWS:  &config.AWSConfig{},
		Name: "correct",
		Contexts: []config.Context{
			config.Context{
				Name: "cluster1",
				Stacks: []config.Stack{
					config.Stack{
						State: &config.StackState{},
					},
					config.Stack{
						Network: &config.StackNetwork{
							NetworkCIDR: "1.2.0.0/20",
						},
					},
					config.Stack{
						Tools: &config.StackTools{},
					},
					config.Stack{
						Vault: &config.StackVault{},
					},
				},
			},
			config.Context{
				Name: "cluster2",
				Stacks: []config.Stack{
					config.Stack{
						Network: &config.StackNetwork{
							NetworkCIDR: "1.2.16.0/20",
						},
					},
				},
			},
			config.Context{
				Name: "cluster3",
				Stacks: []config.Stack{
					config.Stack{
						Network: &config.StackNetwork{
							NetworkCIDR: "1.2.32.0/20",
						},
					},
				},
			},
		},
	}
}

type fakeEnvironment struct {
	*Environment
	ctrl *gomock.Controller

	configPath string
	fakeTarmak *mocks.MockTarmak
}

func (f *fakeEnvironment) Finish() {
	f.ctrl.Finish()
	os.RemoveAll(f.configPath)
}

func newFakeEnvironment(t *testing.T) *fakeEnvironment {

	e := &fakeEnvironment{
		ctrl: gomock.NewController(t),
		Environment: &Environment{
			conf: &config.Environment{
				Name: "fake",
			},
			log: logrus.WithField("test", true),
		},
	}
	e.fakeTarmak = mocks.NewMockTarmak(e.ctrl)
	e.Environment.tarmak = e.fakeTarmak

	// setup custom config path
	var err error
	e.configPath, err = ioutil.TempDir("", "tarmak-fake-env")
	if err != nil {
		t.Fatal("error creating config path: ", err)
	}
	e.fakeTarmak.EXPECT().ConfigPath().AnyTimes().Return(e.configPath)

	// setup custom logger
	logger := logrus.New()
	if testing.Verbose() {
		logger.Level = logrus.DebugLevel
	} else {
		logger.Out = ioutil.Discard
	}
	e.fakeTarmak.EXPECT().Log().AnyTimes().Return(logger.WithField("app", "tarmak"))

	return e
}

func TestNewFromConfigValid(t *testing.T) {
	fakes := newFakeEnvironment(t)
	defer fakes.Finish()
	_, err := NewFromConfig(fakes.fakeTarmak, testDefaultEnvironmentConfig())

	if err != nil {
		t.Error("unexpected error: ", err)
	}
}

func TestNewFromConfigNetworkClash(t *testing.T) {
	fakes := newFakeEnvironment(t)
	defer fakes.Finish()

	cfg := testDefaultEnvironmentConfig()
	cfg.Contexts[1].Stacks[0].Network.NetworkCIDR = "1.2.4.0/24"

	_, err := NewFromConfig(fakes.fakeTarmak, cfg)

	if err == nil {
		t.Error("expect error")
	} else if contain := "network '1.2.0.0/20' overlaps with '1.2.4.0/24'"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}

func TestNewFromConfigMissingState(t *testing.T) {
	fakes := newFakeEnvironment(t)
	defer fakes.Finish()

	cfg := testDefaultEnvironmentConfig()
	cfg.Contexts[0].Stacks = cfg.Contexts[0].Stacks[1:]

	_, err := NewFromConfig(fakes.fakeTarmak, cfg)

	if err == nil {
		t.Error("expect error")
	} else if contain := "has no state stack"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}

func TestNewFromConfigDuplicateState(t *testing.T) {
	fakes := newFakeEnvironment(t)
	defer fakes.Finish()

	cfg := testDefaultEnvironmentConfig()
	cfg.Contexts[0].Stacks = append(cfg.Contexts[0].Stacks, config.Stack{State: &config.StackState{}})

	_, err := NewFromConfig(fakes.fakeTarmak, cfg)

	if err == nil {
		t.Error("expect error")
	} else if contain := "has multiple state stacks"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}

func TestNewFromConfigMissingTools(t *testing.T) {
	fakes := newFakeEnvironment(t)
	defer fakes.Finish()

	cfg := testDefaultEnvironmentConfig()
	stacks := cfg.Contexts[0].Stacks[0:2]
	stacks = append(stacks, cfg.Contexts[0].Stacks[3:]...)
	cfg.Contexts[0].Stacks = stacks

	_, err := NewFromConfig(fakes.fakeTarmak, cfg)

	if err == nil {
		t.Error("expect error")
	} else if contain := "has no tools stack"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}

func TestNewFromConfigDuplicateTools(t *testing.T) {
	fakes := newFakeEnvironment(t)
	defer fakes.Finish()

	cfg := testDefaultEnvironmentConfig()
	cfg.Contexts[0].Stacks = append(cfg.Contexts[0].Stacks, config.Stack{Tools: &config.StackTools{}})

	_, err := NewFromConfig(fakes.fakeTarmak, cfg)

	if err == nil {
		t.Error("expect error")
	} else if contain := "has multiple tools stacks"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}

func TestNewFromConfigMissingVault(t *testing.T) {
	fakes := newFakeEnvironment(t)
	defer fakes.Finish()

	cfg := testDefaultEnvironmentConfig()
	stacks := cfg.Contexts[0].Stacks[0:3]
	stacks = append(stacks, cfg.Contexts[0].Stacks[4:]...)
	cfg.Contexts[0].Stacks = stacks

	_, err := NewFromConfig(fakes.fakeTarmak, cfg)

	if err == nil {
		t.Error("expect error")
	} else if contain := "has no vault stack"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}

func TestNewFromConfigDuplicateVault(t *testing.T) {
	fakes := newFakeEnvironment(t)
	defer fakes.Finish()

	cfg := testDefaultEnvironmentConfig()
	cfg.Contexts[0].Stacks = append(cfg.Contexts[0].Stacks, config.Stack{Vault: &config.StackVault{}})

	_, err := NewFromConfig(fakes.fakeTarmak, cfg)

	if err == nil {
		t.Error("expect error")
	} else if contain := "has multiple vault stacks"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}

func TestEnvironment_SSHPrivateKeyNotExisting(t *testing.T) {
	e := newFakeEnvironment(t)
	defer e.ctrl.Finish()

	// try if we are able to use key as signer
	key := e.SSHPrivateKey()
	_, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal("unable to generate signer out of key: ", err)
	}
}

func TestEnvironment_SSHPrivateKeyExisting(t *testing.T) {
	e := newFakeEnvironment(t)
	defer e.ctrl.Finish()

	err := utils.EnsureDirectory(filepath.Dir(e.SSHPrivateKeyPath()), 0700)
	if err != nil {
		t.Fatal("unable to create dir: ", err)
	}

	err = ioutil.WriteFile(e.SSHPrivateKeyPath(), []byte(fakeSSHKeyInsecurePrivate), 0600)
	if err != nil {
		t.Fatal("unable to write file: ", err)
	}

	// try if we are able to use key as signer
	key := e.SSHPrivateKey()
	_, err = ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal("unable to generate signer out of key: ", err)
	}
}

func TestEnvironment_SSHPrivateKeyExistingGarbage(t *testing.T) {
	e := newFakeEnvironment(t)
	defer e.ctrl.Finish()

	err := utils.EnsureDirectory(filepath.Dir(e.SSHPrivateKeyPath()), 0700)
	if err != nil {
		t.Fatal("unable to create dir: ", err)
	}

	err = ioutil.WriteFile(e.SSHPrivateKeyPath(), []byte("IAMNOTASSHKEY"), 0600)
	if err != nil {
		t.Fatal("unable to write file: ", err)
	}

	// try if we are able to use key as signer
	_, err = e.getSSHPrivateKey()
	if err == nil {
	} else if !strings.Contains(err.Error(), "no key found") {
		t.Error("unexpected error message: ", err)
	}
}
