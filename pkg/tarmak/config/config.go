package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type Config struct {
	tarmak interfaces.Tarmak

	conf *tarmakv1alpha1.Config

	scheme *runtime.Scheme
	codecs serializer.CodecFactory
	log    *logrus.Entry
}

var _ interfaces.Config = &Config{}

func New(tarmak interfaces.Tarmak) (*Config, error) {
	c := &Config{
		tarmak: tarmak,
		log:    tarmak.Log().WithField("module", "config"),
		scheme: runtime.NewScheme(),
	}
	c.codecs = serializer.NewCodecFactory(c.scheme)

	if err := clusterv1alpha1.AddToScheme(c.scheme); err != nil {
		return nil, err
	}

	if err := tarmakv1alpha1.AddToScheme(c.scheme); err != nil {
		return nil, err
	}

	return c, nil
}

func newConfig() *tarmakv1alpha1.Config {
	c := &tarmakv1alpha1.Config{}
	return c
}

func (c *Config) NewAWSConfigClusterSingle() *tarmakv1alpha1.Config {
	conf := newConfig()
	conf.Clusters = []clusterv1alpha1.Cluster{
		*NewClusterSingle("dev", "cluster"),
	}
	provider := NewAWSProfileProvider("dev", "jetstack-dev")
	cluster := NewClusterSingle("dev", "cluster")
	cluster.CloudId = provider.ObjectMeta.Name
	conf.Providers = []tarmakv1alpha1.Provider{*provider}
	conf.Clusters = []clusterv1alpha1.Cluster{*cluster}
	c.scheme.Default(conf)
	return conf
}

func (c *Config) writeYAML(config *tarmakv1alpha1.Config) error {
	var encoder runtime.Encoder
	mediaTypes := c.codecs.SupportedMediaTypes()
	for _, info := range mediaTypes {
		if info.MediaType == "application/yaml" {
			encoder = info.Serializer
			break
		}
	}
	if encoder == nil {
		return errors.New("unable to locate yaml encoder")
	}
	encoder = json.NewYAMLSerializer(json.DefaultMetaFactory, c.scheme, c.scheme)
	encoder = c.codecs.EncoderForVersion(encoder, tarmakv1alpha1.SchemeGroupVersion)

	file, err := os.Create(c.configPath())
	if err != nil {
		return err
	}
	defer file.Close()

	if err := encoder.Encode(config, file); err != nil {
		return err
	}

	return nil
}

func (c *Config) CurrentContextName() string {
	split := strings.Split(c.conf.CurrentContext, "-")
	if len(split) < 2 {
		return ""
	}
	return split[1]
}

func (c *Config) CurrentEnvironmentName() string {
	split := strings.Split(c.conf.CurrentContext, "-")
	return split[0]
}

func (c *Config) Context(environment string, name string) (context *clusterv1alpha1.Cluster, err error) {
	for pos, _ := range c.conf.Clusters {
		context := &c.conf.Clusters[pos]
		if context.Environment == environment && context.Name == name {
			return context, nil
		}
	}
	return nil, fmt.Errorf("context '%s' in environment '%s' not found", name, environment)
}

func (c *Config) Contexts(environment string) (contexts []*clusterv1alpha1.Cluster) {
	for pos, _ := range c.conf.Clusters {
		context := &c.conf.Clusters[pos]
		if context.Environment == environment {
			contexts = append(contexts, context)
		}
	}
	return contexts
}

func (c *Config) Environment(name string) (*tarmakv1alpha1.Environment, error) {
	for pos, _ := range c.conf.Environments {
		environment := &c.conf.Environments[pos]
		if environment.Name == name {
			return environment, nil
		}
	}
	return nil, fmt.Errorf("environment '%s' not found", name)
}

func (c *Config) Environments() (environments []*tarmakv1alpha1.Environment) {
	for pos, _ := range c.conf.Environments {
		environments = append(environments, &c.conf.Environments[pos])
	}
	return environments
}

func (c *Config) Provider(name string) (context *tarmakv1alpha1.Provider, err error) {
	for pos, _ := range c.conf.Providers {
		provider := &c.conf.Providers[pos]
		if provider.Name == name {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("provider '%s' not found", name)
}

func (c *Config) Providers() (providers []*tarmakv1alpha1.Provider) {
	for pos, _ := range c.conf.Providers {
		providers = append(providers, &c.conf.Providers[pos])
	}
	return providers
}

func (c *Config) configPath() string {
	return filepath.Join(c.tarmak.ConfigPath(), "tarmak.yaml")
}

func (c *Config) ReadConfig() (*tarmakv1alpha1.Config, error) {
	path := c.configPath()

	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	configObj, gvk, err := c.codecs.UniversalDecoder(tarmakv1alpha1.SchemeGroupVersion).Decode(configBytes, nil, nil)
	if err != nil {
		return nil, err
	}

	config, ok := configObj.(*tarmakv1alpha1.Config)
	if !ok {
		return nil, fmt.Errorf("got unexpected config type: %v", gvk)
	}

	c.conf = config
	return config, nil
}

func (c *Config) Contact() string {
	return c.conf.Contact
}

func (c *Config) Project() string {
	return c.conf.Project
}
