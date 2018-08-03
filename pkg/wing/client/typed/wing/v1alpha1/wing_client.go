// Copyright Jetstack Ltd. See LICENSE for details.
package v1alpha1

import (
	v1alpha1 "github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	"github.com/jetstack/tarmak/pkg/wing/client/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type WingV1alpha1Interface interface {
	RESTClient() rest.Interface
	InstancesGetter
	MachinesGetter
	PuppetTargetsGetter
	WingJobsGetter
}

// WingV1alpha1Client is used to interact with features provided by the wing.tarmak.io group.
type WingV1alpha1Client struct {
	restClient rest.Interface
}

func (c *WingV1alpha1Client) Instances(namespace string) InstanceInterface {
	return newInstances(c, namespace)
}

func (c *WingV1alpha1Client) Machines(namespace string) MachineInterface {
	return newMachines(c, namespace)
}

func (c *WingV1alpha1Client) PuppetTargets(namespace string) PuppetTargetInterface {
	return newPuppetTargets(c, namespace)
}

func (c *WingV1alpha1Client) WingJobs(namespace string) WingJobInterface {
	return newWingJobs(c, namespace)
}

// NewForConfig creates a new WingV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*WingV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &WingV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new WingV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *WingV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new WingV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *WingV1alpha1Client {
	return &WingV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *WingV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
