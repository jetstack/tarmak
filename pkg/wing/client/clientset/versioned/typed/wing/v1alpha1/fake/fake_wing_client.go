// Copyright Jetstack Ltd. See LICENSE for details.
package fake

import (
	v1alpha1 "github.com/jetstack/tarmak/pkg/wing/client/clientset/versioned/typed/wing/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeWingV1alpha1 struct {
	*testing.Fake
}

func (c *FakeWingV1alpha1) Machines(namespace string) v1alpha1.MachineInterface {
	return &FakeMachines{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeWingV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
