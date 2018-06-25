// Copyright Jetstack Ltd. See LICENSE for details.
package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

type Config struct {
	tarmak interfaces.Tarmak

	conf  *tarmakv1alpha1.Config
	flags *tarmakv1alpha1.Flags

	scheme *runtime.Scheme
	codecs serializer.CodecFactory
	log    *logrus.Entry
}

var _ interfaces.Config = &Config{}

func New(tarmak interfaces.Tarmak, flags *tarmakv1alpha1.Flags) (*Config, error) {
	c := &Config{
		tarmak: tarmak,
		log:    tarmak.Log().WithField("module", "config"),
		flags:  flags,
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

func (c *Config) NewAmazonConfigClusterSingle() *tarmakv1alpha1.Config {
	conf := newConfig()
	conf.Clusters = []clusterv1alpha1.Cluster{
		*NewClusterSingle("dev", "cluster"),
	}
	provider := NewAmazonProfileProvider("dev", "jetstack-dev")
	cluster := NewClusterSingle("dev", "cluster")
	cluster.CloudId = provider.ObjectMeta.Name
	conf.Providers = []tarmakv1alpha1.Provider{*provider}
	conf.Clusters = []clusterv1alpha1.Cluster{*cluster}
	c.scheme.Default(conf)
	return conf
}

func ApplyDefaults(src runtime.Object) error {
	scheme := runtime.NewScheme()

	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}

	if err := tarmakv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}

	scheme.Default(src)
	return nil
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

	err := utils.EnsureDirectory(filepath.Dir(c.configPath()), 0750)
	if err != nil {
		return err
	}

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

func (c *Config) CurrentCluster() (string, error) {
	// override current cluster if flags are set accordingly
	currentCluster := ""
	if c.flags.CurrentCluster != "" {
		currentCluster = c.flags.CurrentCluster
	} else {
		if c.conf == nil {
			return "", fmt.Errorf("config is nil")
		}

		currentCluster = c.conf.CurrentCluster
	}

	split := strings.Split(currentCluster, "-")
	if len(split) != 2 {
		return "", fmt.Errorf("current cluster '%s' is not in the correct format: <environment>-<clusterName>", currentCluster)
	}

	return currentCluster, nil
}

func (c *Config) SetCurrentCluster(clusterName string) error {
	c.conf.CurrentCluster = clusterName
	return c.writeYAML(c.conf)
}

func (c *Config) CurrentClusterName() (string, error) {

	currentCluster, err := c.CurrentCluster()
	if err != nil {
		return "", fmt.Errorf("error retrieving current cluster: %s", err)
	}

	split := strings.Split(currentCluster, "-")
	if len(split) < 2 {
		return "", nil
	}
	return split[1], nil
}

func (c *Config) CurrentEnvironmentName() (string, error) {

	currentCluster, err := c.CurrentCluster()
	if err != nil {
		return "", fmt.Errorf("error retrieving current cluster: %s", err)
	}

	split := strings.Split(currentCluster, "-")
	return split[0], nil
}

func (c *Config) Cluster(environment string, name string) (cluster *clusterv1alpha1.Cluster, err error) {
	for pos, _ := range c.conf.Clusters {
		cluster := &c.conf.Clusters[pos]
		if cluster.Environment == environment && cluster.Name == name {
			return cluster, nil
		}
	}
	return nil, fmt.Errorf("cluster '%s' in environment '%s' not found", name, environment)
}

func (c *Config) Clusters(environment string) (clusters []*clusterv1alpha1.Cluster) {
	for pos, _ := range c.conf.Clusters {
		cluster := &c.conf.Clusters[pos]
		if cluster.Environment == environment {
			clusters = append(clusters, cluster)
		}
	}
	return clusters
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
	if c.conf == nil {
		return environments
	}
	for pos, _ := range c.conf.Environments {
		environments = append(environments, &c.conf.Environments[pos])
	}
	return environments
}

func (c *Config) Provider(name string) (cluster *tarmakv1alpha1.Provider, err error) {
	for pos, _ := range c.conf.Providers {
		provider := &c.conf.Providers[pos]
		if provider.Name == name {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("provider '%s' not found", name)
}

func (c *Config) Providers() (providers []*tarmakv1alpha1.Provider) {
	if c.conf == nil {
		return providers
	}
	for pos, _ := range c.conf.Providers {
		providers = append(providers, &c.conf.Providers[pos])
	}
	return providers
}

func (c *Config) ValidName(name, regex string) error {
	r := regexp.MustCompile(regex)
	str := r.FindString(name)
	if str != name {
		return fmt.Errorf("error matching name '%s' against regex %s", name, regex)
	}

	return nil
}

func (c *Config) UniqueProviderName(name string) error {
	for _, p := range c.Providers() {
		if p.Name == name {
			return fmt.Errorf("name '%s' not unique", name)
		}
	}
	return nil
}

func (c *Config) AppendProvider(prov *tarmakv1alpha1.Provider) error {
	if c.conf == nil {
		c.conf = &tarmakv1alpha1.Config{}
		c.scheme.Default(c.conf)
	}

	if err := c.UniqueProviderName(prov.Name); err != nil {
		return fmt.Errorf("failed to add provider: %v", err)
	}

	c.scheme.Default(prov)
	c.conf.Providers = append(c.conf.Providers, *prov)
	return c.writeYAML(c.conf)
}

func (c *Config) AppendEnvironment(env *tarmakv1alpha1.Environment) error {
	if c.conf == nil {
		c.conf = &tarmakv1alpha1.Config{}
		c.scheme.Default(c.conf)
	}

	if err := c.UniqueEnvironmentName(env.Name); err != nil {
		return fmt.Errorf("failed to add environment: %v", err)
	}

	c.scheme.Default(env)
	c.conf.Environments = append(c.conf.Environments, *env)
	return c.writeYAML(c.conf)
}

func (c *Config) UniqueEnvironmentName(name string) error {
	for _, e := range c.Environments() {
		if e.Name == name {
			return fmt.Errorf("name '%s' not unique", name)
		}
	}
	return nil
}

func (c *Config) AppendCluster(cluster *clusterv1alpha1.Cluster) error {
	if c.conf == nil {
		c.conf = &tarmakv1alpha1.Config{}
		c.scheme.Default(c.conf)
	}

	if err := c.UniqueClusterName(cluster.Environment, cluster.Name); err != nil {
		return fmt.Errorf("failed to add cluster: %v", err)
	}

	c.scheme.Default(cluster)
	c.conf.Clusters = append(c.conf.Clusters, *cluster)
	return c.writeYAML(c.conf)
}

func (c *Config) UniqueClusterName(environment, name string) error {
	for _, u := range c.Clusters(environment) {
		if u.Name == name {
			return fmt.Errorf("name '%s' not unique", name)
		}
	}
	return nil
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

	_, err = c.readClusters()
	if err != nil {
		return nil, fmt.Errorf("failed to read cluster configs: %v", err)
	}

	c.conf = config
	return config, nil
}

func (c *Config) readClusters() ([]*clusterv1alpha1.Cluster, error) {
	dir, err := ioutil.ReadDir(c.tarmak.ConfigPath())
	if err != nil {
		return nil, nil
	}

	var clusterFiles []os.FileInfo
	for _, f := range dir {
		if !f.IsDir() && strings.HasPrefix(f.Name(), "cluster") && strings.HasSuffix(f.Name(), ".yaml") {
			clusterFiles = append(clusterFiles, f)
		}
	}

	var result *multierror.Error
	var clusterConfigs []*clusterv1alpha1.Cluster
	for _, f := range clusterFiles {
		b, err := ioutil.ReadFile(filepath.Join(c.tarmak.ConfigPath(), f.Name()))
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		configObj, gvk, err := c.codecs.UniversalDeserializer().Decode(b, nil, nil)
		if err != nil {
			err = fmt.Errorf("failed to decode cluster config: %v", err)
			result = multierror.Append(result, err)
			continue
		}

		clusterConfig, ok := configObj.(*clusterv1alpha1.Cluster)
		if !ok {
			result = multierror.Append(result, fmt.Errorf("got unexpected config type: %v", gvk))
			continue
		}

		clusterConfigs = append(clusterConfigs, clusterConfig)
	}

	return clusterConfigs, result.ErrorOrNil()
}

func (c *Config) Contact() string {
	return c.conf.Contact
}

func (c *Config) Project() string {
	return c.conf.Project
}

func (c *Config) WingDevMode() bool {
	return c.flags.WingDevMode
}
