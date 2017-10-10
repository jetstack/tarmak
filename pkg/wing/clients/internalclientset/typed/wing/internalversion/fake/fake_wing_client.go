// Copyright Jetstack Ltd. See LICENSE for details.
package fake

import (
	internalversion "github.com/jetstack/tarmak/pkg/wing/clients/internalclientset/typed/wing/internalversion"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeWing struct {
	*testing.Fake
}

func (c *FakeWing) Instances(namespace string) internalversion.InstanceInterface {
	return &FakeInstances{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeWing) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
