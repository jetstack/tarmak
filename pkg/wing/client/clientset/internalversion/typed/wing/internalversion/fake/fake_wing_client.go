// Copyright Jetstack Ltd. See LICENSE for details.
package fake

import (
	internalversion "github.com/jetstack/tarmak/pkg/wing/client/clientset/internalversion/typed/wing/internalversion"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeWing struct {
	*testing.Fake
}

func (c *FakeWing) Machines(namespace string) internalversion.MachineInterface {
	return &FakeMachines{c, namespace}
}

func (c *FakeWing) MachineDeployments(namespace string) internalversion.MachineDeploymentInterface {
	return &FakeMachineDeployments{c, namespace}
}

func (c *FakeWing) MachineSets(namespace string) internalversion.MachineSetInterface {
	return &FakeMachineSets{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeWing) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
