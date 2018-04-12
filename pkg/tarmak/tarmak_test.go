// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/mocks"
)

type testTarmak struct {
	ctrl *gomock.Controller

	logger *logrus.Logger

	tarmak *Tarmak

	// temporary config directory
	configDirectory string
	environments    []*tarmakv1alpha1.Environment
	clusters        []*clusterv1alpha1.Cluster

	fakeConfig   *mocks.MockConfig
	fakeProvider *mocks.MockProvider
}

func (tt *testTarmak) finish() {
	if tt.configDirectory != "" {
		err := os.RemoveAll(tt.configDirectory)
		if err != nil {
			tt.logger.Warn("error deleting config directory: ", err)
		}
	}
	tt.ctrl.Finish()
}

func (tt *testTarmak) fakeAWSProvider(name string) {
	baseImage := tarmakv1alpha1.Image{}
	baseImage.Name = "ami-6e28b517"

	tt.fakeProvider.EXPECT().Name().AnyTimes().Return(name)
	tt.fakeProvider.EXPECT().Cloud().AnyTimes().Return("amazon")
	tt.fakeProvider.EXPECT().InstanceType(gomock.Any()).AnyTimes().Return("t2.large", nil)
	tt.fakeProvider.EXPECT().VolumeType(gomock.Any()).AnyTimes().Return("ssd", nil)
	tt.fakeProvider.EXPECT().Validate().AnyTimes().Return(nil)
	tt.fakeProvider.EXPECT().RemoteState(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("")
	tt.fakeProvider.EXPECT().RemoteStateBucketName().AnyTimes().Return("my-remote-bucket")
	tt.fakeProvider.EXPECT().QueryImages(gomock.Any()).AnyTimes().Return([]tarmakv1alpha1.Image{baseImage}, nil)
	tt.fakeProvider.EXPECT().Variables().AnyTimes().Return(map[string]interface{}{
		"test": "ffs",
	})

	// override provider creation method
	tt.tarmak.providerByName = func(providerName string) (interfaces.Provider, error) {
		return tt.fakeProvider, nil
	}

	tt.fakeConfig.EXPECT().Provider(name).AnyTimes().Return(&tarmakv1alpha1.Provider{}, nil)
}

func (tt *testTarmak) addEnvironment(env *tarmakv1alpha1.Environment) {
	tt.environments = append(tt.environments, env)
	tt.fakeConfig.EXPECT().Environment(env.Name).Return(env, nil)
	tt.fakeConfig.EXPECT().CurrentEnvironmentName().Return(env.Name)
	tt.fakeConfig.EXPECT().Environments().AnyTimes().Return(tt.environments)
}

func (tt *testTarmak) addCluster(cluster *clusterv1alpha1.Cluster) {
	tt.clusters = append(tt.clusters, cluster)
	tt.fakeConfig.EXPECT().Cluster(cluster.Environment, cluster.Name).AnyTimes().Return(cluster, nil)
	tt.fakeConfig.EXPECT().CurrentClusterName().Return(cluster.Name)
	// TODO: support multiple environments
	tt.fakeConfig.EXPECT().Clusters(cluster.Environment).AnyTimes().Return(tt.clusters)
}

func newTestTarmak(t *testing.T) *testTarmak {

	logger := logrus.New()
	if testing.Verbose() {
		logger.Level = logrus.DebugLevel
	} else {
		logger.Out = ioutil.Discard
	}

	tt := &testTarmak{
		ctrl: gomock.NewController(t),
		tarmak: &Tarmak{
			log:   logger,
			flags: &tarmakv1alpha1.Flags{},
		},
		logger: logger,
	}

	var err error
	if tt.configDirectory, err = ioutil.TempDir("", "tarmak-test"); err != nil {
		t.Fatal("error creating temporary config directory", err)
	}
	tt.tarmak.configDirectory = tt.configDirectory
	tt.tarmak.flags.ConfigDirectory = tt.configDirectory
	tt.logger.WithField("config_directory", tt.configDirectory).Debug("created temporary config folder")

	tt.tarmak.initializeModules()

	tt.fakeConfig = mocks.NewMockConfig(tt.ctrl)
	tt.fakeConfig.EXPECT().Contact().AnyTimes().Return("tech+testing@jetstack.io")
	tt.fakeConfig.EXPECT().Project().AnyTimes().Return("testing")

	tt.fakeProvider = mocks.NewMockProvider(tt.ctrl)
	tt.tarmak.config = tt.fakeConfig

	return tt

}

func newTestTarmakClusterSingle(t *testing.T) *testTarmak {
	tt := newTestTarmak(t)

	tt.fakeAWSProvider("aws")

	env := config.NewEnvironment("single", "test", "tech+test@jetstack.io")
	env.Provider = "aws"
	tt.addEnvironment(env)
	tt.addCluster(config.NewClusterSingle(env.Name, "cluster"))

	if err := tt.tarmak.initializeConfig(); err != nil {
		t.Fatal("error intializing tarmak: ", err)
	}

	return tt
}

func newTestTarmakClusterMulti(t *testing.T) *testTarmak {

	tt := newTestTarmak(t)

	tt.fakeAWSProvider("aws")

	env := config.NewEnvironment("multi", "test", "tech+test@jetstack.io")
	env.Provider = "aws"
	tt.addEnvironment(env)
	tt.addCluster(config.NewClusterMulti(env.Name, "test"))

	if err := tt.tarmak.initializeConfig(); err != nil {
		t.Fatal("error intializing tarmak: ", err)
	}

	return tt
}

func newTestTarmakHub(t *testing.T) *testTarmak {
	tt := newTestTarmak(t)

	tt.fakeAWSProvider("aws")

	env := config.NewEnvironment("multi", "test", "tech+test@jetstack.io")
	env.Provider = "aws"
	tt.addEnvironment(env)
	tt.addCluster(config.NewHub(env.Name))

	if err := tt.tarmak.initializeConfig(); err != nil {
		t.Fatal("error intializing tarmak: ", err)
	}

	return tt
}

func TestTarmak_Terraform_Generate_ClusterSingle(t *testing.T) {
	tt := newTestTarmakClusterSingle(t)
	defer tt.finish()
	tarmak := tt.tarmak

	if err := tarmak.Validate(); err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if err := tarmak.terraform.GenerateCode(tarmak.Cluster()); err != nil {
		t.Fatal("Unexpected error:", err)
	}

	tt.logger.WithField("config_path", tt.tarmak.ConfigPath()).Debug("created temporary config folder")

}

func TestTarmak_Terraform_Generate_ClusterMulti(t *testing.T) {
	tt := newTestTarmakClusterMulti(t)
	defer tt.finish()
	tarmak := tt.tarmak

	if err := tarmak.Validate(); err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if err := tarmak.terraform.GenerateCode(tarmak.Cluster()); err != nil {
		t.Fatal("Unexpected error:", err)
	}

	tt.logger.WithField("config_path", tt.tarmak.ConfigPath()).Debug("created temporary config folder")

}

func TestTarmak_Terraform_Generate_Hub(t *testing.T) {
	tt := newTestTarmakHub(t)
	defer tt.finish()
	tarmak := tt.tarmak

	if err := tarmak.Validate(); err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if err := tarmak.terraform.GenerateCode(tarmak.Cluster()); err != nil {
		t.Fatal("Unexpected error:", err)
	}

	tt.logger.WithField("config_path", tt.tarmak.ConfigPath()).Debug("created temporary config folder")

}
