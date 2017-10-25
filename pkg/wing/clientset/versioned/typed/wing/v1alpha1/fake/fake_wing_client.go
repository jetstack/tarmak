// Copyright Jetstack Ltd. See LICENSE for details.
package fake

import (
	v1alpha1 "github.com/jetstack/tarmak/pkg/wing/clientset/versioned/typed/wing/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeWingV1alpha1 struct {
	*testing.Fake
}

func (c *FakeWingV1alpha1) Instances(namespace string) v1alpha1.InstanceInterface {
	return &FakeInstances{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeWingV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
