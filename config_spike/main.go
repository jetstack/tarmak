package main

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime"
	yamlUtil "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	//"github.com/jetstack/tarmak/pkg/client"
	"github.com/jetstack/tarmak/pkg/client/scheme"
)

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

func main() {
	configPath := "/home/christian/.tarmak/tarmak.yaml"

	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}

	decoder := yamlUtil.NewYAMLOrJSONDecoder(file, 1024*1024)

	conf := v1alpha1.Config{}

	err = decoder.Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	confObjNew, err := convert(runtime.Object(&conf))
	if err != nil {
		log.Fatal(err)
	}
	confNew := confObjNew.(*v1alpha1.Config)

	output, err := yaml.Marshal(*confNew)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("config:\n%s\n", string(output))
}
