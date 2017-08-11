package node_group

import (
	"testing"

	"github.com/jetstack/tarmak/pkg/tarmak/mocks"
)

func testDefaultContextConfig() *config.Context {
	return &config.Context{
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
	}
}

type fakeContext struct {
	*Context
	ctrl *gomock.Controller

	fakeEnvironment *mocks.MockEnvironment
}

func (f *fakeContext) Finish() {
	f.ctrl.Finish()
}

func newFakeContext(t *testing.T) *fakeContext {

	c := &fakeContext{
		ctrl: gomock.NewController(t),
		Context: &Context{
			conf: testDefaultContextConfig(),
		},
	}
	c.fakeEnvironment = mocks.NewMockEnvironment(c.ctrl)
	c.Context.environment = c.fakeEnvironment

	// setup custom logger
	logger := logrus.New()
	loggerCtx := logger.WithField("app", "tarmak")
	if testing.Verbose() {
		logger.Level = logrus.DebugLevel
	} else {
		logger.Out = ioutil.Discard
	}
	c.fakeEnvironment.EXPECT().Log().AnyTimes().Return(loggerCtx)
	c.Context.log = loggerCtx

	return c
}

func TestNewFromConfigValid(t *testing.T) {
	fakes := newFakeContext(t)
	defer fakes.Finish()
	_, err := NewFromConfig(fakes.fakeEnvironment, testDefaultContextConfig())

	if err != nil {
		t.Error("unexpected error: ", err)
	}
}

func TestNewFromConfigInvalidNetwork(t *testing.T) {
	fakes := newFakeContext(t)
	defer fakes.Finish()

	cfg := testDefaultContextConfig()
	cfg.Stacks[1].Network.NetworkCIDR = "260.0.2.0/24"
	_, err := NewFromConfig(fakes.fakeEnvironment, cfg)

	if err == nil {
		t.Error("expect error")
	} else if contain := "invalid CIDR address"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}

func TestNewFromConfigMissingNetwork(t *testing.T) {
	fakes := newFakeContext(t)
	defer fakes.Finish()

	cfg := testDefaultContextConfig()
	stacks := cfg.Stacks[0:1]
	stacks = append(stacks, cfg.Stacks[2:]...)
	cfg.Stacks = stacks

	_, err := NewFromConfig(fakes.fakeEnvironment, cfg)

	if err == nil {
		t.Error("expect error")
	} else if contain := "has no network stack"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}

func TestNewFromConfigDuplicateNetwork(t *testing.T) {
	fakes := newFakeContext(t)
	defer fakes.Finish()

	cfg := testDefaultContextConfig()
	cfg.Stacks = append(cfg.Stacks, config.Stack{Network: &config.StackNetwork{NetworkCIDR: "1.2.4.0/22"}})

	_, err := NewFromConfig(fakes.fakeEnvironment, cfg)

	if err == nil {
		t.Error("expect error")
	} else if contain := "has multiple network stacks"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}
