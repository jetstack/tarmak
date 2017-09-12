package context

import (
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/golang/mock/gomock"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/mocks"
)

type fakeContext struct {
	*Context
	ctrl *gomock.Controller

	fakeEnvironment *mocks.MockEnvironment
	fakeProvider    *mocks.MockProvider
	fakeTarmak      *mocks.MockTarmak
	fakeConfig      *mocks.MockConfig
}

func (f *fakeContext) Finish() {
	f.ctrl.Finish()
}

func newFakeContext(t *testing.T, cluster *clusterv1alpha1.Cluster) *fakeContext {

	c := &fakeContext{
		ctrl: gomock.NewController(t),
		Context: &Context{
			conf: cluster,
		},
	}
	c.fakeEnvironment = mocks.NewMockEnvironment(c.ctrl)
	c.fakeProvider = mocks.NewMockProvider(c.ctrl)
	c.fakeTarmak = mocks.NewMockTarmak(c.ctrl)
	c.fakeConfig = mocks.NewMockConfig(c.ctrl)
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
	c.fakeEnvironment.EXPECT().Provider().AnyTimes().Return(c.fakeProvider)
	c.fakeEnvironment.EXPECT().Tarmak().AnyTimes().Return(c.fakeTarmak)
	c.Context.log = loggerCtx

	c.fakeProvider.EXPECT().InstanceType(gomock.Any()).Do(func(in string) string { return "provider-" + in }).AnyTimes()
	c.fakeProvider.EXPECT().VolumeType(gomock.Any()).Do(func(in string) string { return "provider-" + in }).AnyTimes()
	c.fakeProvider.EXPECT().Name().Return("provider").AnyTimes()

	c.fakeTarmak.EXPECT().Config().AnyTimes().Return(c.fakeConfig)

	return c
}

func TestContext_NewMinimalClusterMulti(t *testing.T) {
	clusterConfig := config.NewClusterMulti("multi", "cluster")
	clusterConfig.Location = "my-region"
	c := newFakeContext(t, nil)

	// fake two contexts
	c.fakeEnvironment.EXPECT().Contexts().AnyTimes().Return([]interfaces.Context{mocks.NewMockContext(c.ctrl), mocks.NewMockContext(c.ctrl)})
	c.fakeEnvironment.EXPECT().Name().Return("multi").AnyTimes()

	var err error
	c.Context, err = NewFromConfig(c.fakeEnvironment, clusterConfig)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	if act, exp := c.Name(), "cluster"; act != exp {
		t.Errorf("unexpected name, actual = '%s', expected = '%s'", act, exp)
	}

	if act, exp := c.Type(), "cluster-multi"; act != exp {
		t.Errorf("unexpected type, actual = '%s', expected = '%s'", act, exp)
	}

	if act, exp := c.Region(), "my-region"; act != exp {
		t.Errorf("unexpected region, actual = '%s', expected = '%s'", act, exp)
	}
}

func TestContext_NewMinimalClusterSingle(t *testing.T) {
	clusterConfig := config.NewClusterSingle("single", "cluster")
	clusterConfig.Location = "my-region"
	c := newFakeContext(t, nil)

	// fake single context
	c.fakeEnvironment.EXPECT().Contexts().AnyTimes().Return([]interfaces.Context{mocks.NewMockContext(c.ctrl)})

	var err error
	c.Context, err = NewFromConfig(c.fakeEnvironment, clusterConfig)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	if act, exp := c.Name(), "cluster"; act != exp {
		t.Errorf("unexpected name, actual = '%s', expected = '%s'", act, exp)
	}

	if act, exp := c.Type(), "cluster-single"; act != exp {
		t.Errorf("unexpected type, actual = '%s', expected = '%s'", act, exp)
	}

	if act, exp := c.Region(), "my-region"; act != exp {
		t.Errorf("unexpected region, actual = '%s', expected = '%s'", act, exp)
	}
}

func TestContext_NewMinimalHub(t *testing.T) {
	clusterConfig := config.NewHub("multi")
	clusterConfig.Location = "my-region"
	c := newFakeContext(t, nil)

	// fake two contexts
	c.fakeEnvironment.EXPECT().Contexts().AnyTimes().Return([]interfaces.Context{mocks.NewMockContext(c.ctrl), mocks.NewMockContext(c.ctrl)})

	var err error
	c.Context, err = NewFromConfig(c.fakeEnvironment, clusterConfig)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	if act, exp := c.Name(), "hub"; act != exp {
		t.Errorf("unexpected name, actual = '%s', expected = '%s'", act, exp)
	}

	if act, exp := c.Type(), "hub"; act != exp {
		t.Errorf("unexpected type, actual = '%s', expected = '%s'", act, exp)
	}

	if act, exp := c.Region(), "my-region"; act != exp {
		t.Errorf("unexpected region, actual = '%s', expected = '%s'", act, exp)
	}
}

/*
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

*/
