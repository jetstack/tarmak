// Copyright Jetstack Ltd. See LICENSE for details.
package cluster

import (
	"io/ioutil"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/mocks"
)

type fakeCluster struct {
	*Cluster
	ctrl *gomock.Controller

	fakeEnvironment *mocks.MockEnvironment
	fakeProvider    *mocks.MockProvider
	fakeTarmak      *mocks.MockTarmak
	fakeConfig      *mocks.MockConfig
}

func (f *fakeCluster) Finish() {
	f.ctrl.Finish()
}

func newFakeCluster(t *testing.T, cluster *clusterv1alpha1.Cluster) *fakeCluster {

	c := &fakeCluster{
		ctrl: gomock.NewController(t),
		Cluster: &Cluster{
			conf: cluster,
		},
	}

	c.fakeEnvironment = mocks.NewMockEnvironment(c.ctrl)
	c.fakeProvider = mocks.NewMockProvider(c.ctrl)
	c.fakeTarmak = mocks.NewMockTarmak(c.ctrl)
	c.fakeConfig = mocks.NewMockConfig(c.ctrl)
	c.Cluster.environment = c.fakeEnvironment

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
	c.Cluster.log = loggerCtx

	c.fakeProvider.EXPECT().InstanceType(gomock.Any()).Do(func(in string) string { return "provider-" + in }).AnyTimes()
	c.fakeProvider.EXPECT().VolumeType(gomock.Any()).Do(func(in string) string { return "provider-" + in }).AnyTimes()
	c.fakeProvider.EXPECT().Cloud().Return("provider").AnyTimes()
	c.fakeProvider.EXPECT().Name().Return("provider-name").AnyTimes()

	c.fakeTarmak.EXPECT().Config().AnyTimes().Return(c.fakeConfig)

	return c
}

func TestCluster_NewMinimalClusterMulti(t *testing.T) {
	clusterConfig := config.NewClusterMulti("multi", "cluster")
	config.ApplyDefaults(clusterConfig)
	clusterConfig.Location = "my-region"
	c := newFakeCluster(t, nil)
	defer c.Finish()

	// fake two clusters
	c.fakeEnvironment.EXPECT().Name().Return("multi").AnyTimes()
	c.fakeConfig.EXPECT().Clusters("multi").Return([]*clusterv1alpha1.Cluster{
		&clusterv1alpha1.Cluster{},
		&clusterv1alpha1.Cluster{},
	}).AnyTimes()

	var err error
	c.Cluster, err = NewFromConfig(c.fakeEnvironment, clusterConfig)
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

func TestCluster_NewMinimalClusterSingle(t *testing.T) {
	clusterConfig := config.NewClusterSingle("single", "cluster")
	config.ApplyDefaults(clusterConfig)
	clusterConfig.Location = "my-region"
	c := newFakeCluster(t, nil)
	defer c.Finish()

	// fake single cluster
	c.fakeEnvironment.EXPECT().Name().Return("single").AnyTimes()
	c.fakeConfig.EXPECT().Clusters("single").Return([]*clusterv1alpha1.Cluster{
		&clusterv1alpha1.Cluster{},
	}).AnyTimes()

	var err error
	c.Cluster, err = NewFromConfig(c.fakeEnvironment, clusterConfig)
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

func TestCluster_NewMinimalHub(t *testing.T) {
	clusterConfig := config.NewHub("multi")
	config.ApplyDefaults(clusterConfig)
	clusterConfig.Location = "my-region"
	c := newFakeCluster(t, nil)
	defer c.Finish()

	// fake two clusters
	c.fakeEnvironment.EXPECT().Name().Return("single").AnyTimes()
	c.fakeConfig.EXPECT().Clusters("single").Return([]*clusterv1alpha1.Cluster{
		&clusterv1alpha1.Cluster{},
		&clusterv1alpha1.Cluster{},
	}).AnyTimes()

	var err error
	c.Cluster, err = NewFromConfig(c.fakeEnvironment, clusterConfig)
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

func TestCluster_ValidateClusterInstancePoolTypesHub(t *testing.T) {
	clusterConfig := config.NewHub("multi")
	config.ApplyDefaults(clusterConfig)
	clusterConfig.Location = "my-region"
	c := newFakeCluster(t, nil)
	defer c.Finish()

	var err error
	c.Cluster, err = NewFromConfig(c.fakeEnvironment, clusterConfig)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	if err := c.Cluster.validateMultiClusterInstancePoolTypes(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	volume := clusterv1alpha1.Volume{}
	volume.Name = "root"

	for _, i := range []string{
		clusterv1alpha1.InstancePoolTypeVault,
		clusterv1alpha1.InstancePoolTypeBastion,
	} {
		config := clusterConfig
		config.InstancePools = append(config.InstancePools, clusterv1alpha1.InstancePool{
			Type:     i,
			MaxCount: 1,
			Volumes:  []clusterv1alpha1.Volume{volume},
		})

		c.Cluster, err = NewFromConfig(c.fakeEnvironment, config)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
	}

	for _, i := range []string{
		clusterv1alpha1.InstancePoolTypeMaster,
		clusterv1alpha1.InstancePoolTypeWorker,
		clusterv1alpha1.InstancePoolTypeEtcd,
		clusterv1alpha1.InstancePoolTypeJenkins,
		clusterv1alpha1.InstancePoolTypeMasterEtcd,
	} {
		config := clusterConfig
		config.InstancePools = append(config.InstancePools, clusterv1alpha1.InstancePool{
			Type:     i,
			MaxCount: 1,
			Volumes:  []clusterv1alpha1.Volume{volume},
		})

		c.Cluster, err = NewFromConfig(c.fakeEnvironment, config)
		if err == nil {
			t.Errorf("expected error, got=none. type (%s)", i)
		}
	}
}

func TestCluster_ValidateClusterInstancePoolTypesMulti(t *testing.T) {
	clusterConfig := config.NewClusterMulti("multi", "foo")
	config.ApplyDefaults(clusterConfig)
	clusterConfig.Location = "my-region"
	c := newFakeCluster(t, nil)
	defer c.Finish()

	var err error
	c.Cluster, err = NewFromConfig(c.fakeEnvironment, clusterConfig)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	if err := c.Cluster.validateMultiClusterInstancePoolTypes(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	volume := clusterv1alpha1.Volume{}
	volume.Name = "root"

	for _, i := range []string{
		clusterv1alpha1.InstancePoolTypeMaster,
		clusterv1alpha1.InstancePoolTypeWorker,
		clusterv1alpha1.InstancePoolTypeEtcd,
		clusterv1alpha1.InstancePoolTypeJenkins,
		clusterv1alpha1.InstancePoolTypeMasterEtcd,
	} {
		config := clusterConfig
		config.InstancePools = append(config.InstancePools, clusterv1alpha1.InstancePool{
			Type:     i,
			MaxCount: 1,
			Volumes:  []clusterv1alpha1.Volume{volume},
		})

		c.Cluster, err = NewFromConfig(c.fakeEnvironment, config)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
	}

	for _, i := range []string{
		clusterv1alpha1.InstancePoolTypeVault,
		clusterv1alpha1.InstancePoolTypeBastion,
	} {
		config := clusterConfig
		config.InstancePools = append(config.InstancePools, clusterv1alpha1.InstancePool{
			Type:     i,
			MaxCount: 1,
			Volumes:  []clusterv1alpha1.Volume{volume},
		})

		c.Cluster, err = NewFromConfig(c.fakeEnvironment, config)
		if err == nil {
			t.Errorf("expected error, got=none. type (%s)", i)
		}
	}
}

/*
func testDefaultClusterConfig() *config.Cluster {
	return &config.Cluster{
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
	fakes := newFakeCluster(t)
	defer fakes.Finish()
	_, err := NewFromConfig(fakes.fakeEnvironment, testDefaultClusterConfig())

	if err != nil {
		t.Error("unexpected error: ", err)
	}
}

func TestNewFromConfigInvalidNetwork(t *testing.T) {
	fakes := newFakeCluster(t)
	defer fakes.Finish()

	cfg := testDefaultClusterConfig()
	cfg.Stacks[1].Network.NetworkCIDR = "260.0.2.0/24"
	_, err := NewFromConfig(fakes.fakeEnvironment, cfg)

	if err == nil {
		t.Error("expect error")
	} else if contain := "invalid CIDR address"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}

func TestNewFromConfigMissingNetwork(t *testing.T) {
	fakes := newFakeCluster(t)
	defer fakes.Finish()

	cfg := testDefaultClusterConfig()
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
	fakes := newFakeCluster(t)
	defer fakes.Finish()

	cfg := testDefaultClusterConfig()
	cfg.Stacks = append(cfg.Stacks, config.Stack{Network: &config.StackNetwork{NetworkCIDR: "1.2.4.0/22"}})

	_, err := NewFromConfig(fakes.fakeEnvironment, cfg)

	if err == nil {
		t.Error("expect error")
	} else if contain := "has multiple network stacks"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}

*/
