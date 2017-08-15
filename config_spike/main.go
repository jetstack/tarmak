package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/jetstack/tarmak/pkg/apis/tarmak"
	"github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	//"github.com/jetstack/tarmak/pkg/client"
)

type Config struct {
	scheme *runtime.Scheme
	codecs serializer.CodecFactory
}

func NewConfig() (*Config, error) {
	c := &Config{}

	c.scheme = runtime.NewScheme()
	c.codecs = serializer.NewCodecFactory(c.scheme)

	if err := tarmak.AddToScheme(c.scheme); err != nil {
		return nil, err
	}

	if err := v1alpha1.AddToScheme(c.scheme); err != nil {
		return nil, err
	}

	return c, nil

}

func (c *Config) WriteYAML(conf *v1alpha1.Config) error {
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
	encoder = c.codecs.EncoderForVersion(encoder, v1alpha1.SchemeGroupVersion)

	configTo := "test.yaml"
	configFile, err := os.Create(configTo)
	if err != nil {
		return err
	}
	defer configFile.Close()

	if err := encoder.Encode(conf, configFile); err != nil {
		return err
	}

	fmt.Printf("Wrote configuration to: %s\n", configTo)

	return nil
}

func (c *Config) ReadFile(configPath string) (*v1alpha1.Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	configObj, gvk, err := c.codecs.UniversalDecoder(v1alpha1.SchemeGroupVersion).Decode(data, nil, nil)
	if err != nil {
		return nil, err
	}
	config, ok := configObj.(*v1alpha1.Config)
	if !ok {
		return nil, fmt.Errorf("got unexpected config type: %v", gvk)
	}

	return config, nil

}

/*
func convert(obj runtime.Object) (runtime.Object, error) {

	data, err := runtime.Encode(scheme.Codecs.LegacyCodec(v1alpha1.SchemeGroupVersion), obj)
	if err != nil {
		return nil, fmt.Errorf("%v\n %#v", err, obj)
	}
	obj2, err := runtime.Decode(scheme.Codecs.UniversalDecoder(), data)
	if err != nil {
		return nil, fmt.Errorf("%v\nData: %s\nSource: %#v", err, string(data), obj)
	}
	obj3 := reflect.New(reflect.TypeOf(obj).Elem()).Interface().(runtime.Object)
	err = scheme.Scheme.Convert(obj2, obj3, nil)
	if err != nil {
		return nil, fmt.Errorf("%v\nSource: %#v", err, obj2)
	}
	return obj3, nil
}
*/

func main() {
	c, err := NewConfig()
	if err != nil {
		log.Fatal("error init config: ", err)
	}

	//conf := &v1alpha1.Config{}
	conf, err := c.ReadFile("/home/christian/.tarmak/tarmak.yaml")
	if err != nil {
		log.Fatal("error reading config: ", err)
	}

	err = c.WriteYAML(conf)
	if err != nil {
		log.Fatal("error writing config: ", err)
	}
}
