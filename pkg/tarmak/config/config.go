package config

import (
	"errors"
	"io"

	"github.com/Sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
)

var scheme *runtime.Scheme
var codecs serializer.CodecFactory
var log *logrus.Entry

func init() {
	log = logrus.New().WithField("module", "config")

	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)

	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		panic(err)
	}

	if err := tarmakv1alpha1.AddToScheme(scheme); err != nil {
		panic(err)
	}
}

func newConfig() *tarmakv1alpha1.Config {
	c := &tarmakv1alpha1.Config{}
	return c
}

func NewAWSConfigClusterSingle() *tarmakv1alpha1.Config {
	c := newConfig()
	c.Clusters = []clusterv1alpha1.Cluster{
		*NewClusterSingle("dev", "cluster"),
	}
	provider := NewAWSProfileProvider("dev", "jetstack-dev")
	cluster := NewClusterSingle("dev", "cluster")
	cluster.CloudId = provider.ObjectMeta.Name
	c.Providers = []tarmakv1alpha1.Provider{*provider}
	c.Clusters = []clusterv1alpha1.Cluster{*cluster}
	scheme.Default(c)
	return c
}

func writeYAML(config *tarmakv1alpha1.Config, destination io.Writer) error {
	var encoder runtime.Encoder
	mediaTypes := codecs.SupportedMediaTypes()
	for _, info := range mediaTypes {
		if info.MediaType == "application/yaml" {
			encoder = info.Serializer
			break
		}
	}
	if encoder == nil {
		return errors.New("unable to locate yaml encoder")
	}
	encoder = json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)
	encoder = codecs.EncoderForVersion(encoder, tarmakv1alpha1.SchemeGroupVersion)

	if err := encoder.Encode(config, destination); err != nil {
		return err
	}

	return nil
}
