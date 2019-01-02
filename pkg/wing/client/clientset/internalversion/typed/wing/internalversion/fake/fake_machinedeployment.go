// Copyright Jetstack Ltd. See LICENSE for details.
package fake

import (
	wing "github.com/jetstack/tarmak/pkg/apis/wing"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMachineDeployments implements MachineDeploymentInterface
type FakeMachineDeployments struct {
	Fake *FakeWing
	ns   string
}

var machinedeploymentsResource = schema.GroupVersionResource{Group: "wing.tarmak.io", Version: "", Resource: "machinedeployments"}

var machinedeploymentsKind = schema.GroupVersionKind{Group: "wing.tarmak.io", Version: "", Kind: "MachineDeployment"}

// Get takes name of the machineDeployment, and returns the corresponding machineDeployment object, and an error if there is any.
func (c *FakeMachineDeployments) Get(name string, options v1.GetOptions) (result *wing.MachineDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(machinedeploymentsResource, c.ns, name), &wing.MachineDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*wing.MachineDeployment), err
}

// List takes label and field selectors, and returns the list of MachineDeployments that match those selectors.
func (c *FakeMachineDeployments) List(opts v1.ListOptions) (result *wing.MachineDeploymentList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(machinedeploymentsResource, machinedeploymentsKind, c.ns, opts), &wing.MachineDeploymentList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &wing.MachineDeploymentList{}
	for _, item := range obj.(*wing.MachineDeploymentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested machineDeployments.
func (c *FakeMachineDeployments) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(machinedeploymentsResource, c.ns, opts))

}

// Create takes the representation of a machineDeployment and creates it.  Returns the server's representation of the machineDeployment, and an error, if there is any.
func (c *FakeMachineDeployments) Create(machineDeployment *wing.MachineDeployment) (result *wing.MachineDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(machinedeploymentsResource, c.ns, machineDeployment), &wing.MachineDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*wing.MachineDeployment), err
}

// Update takes the representation of a machineDeployment and updates it. Returns the server's representation of the machineDeployment, and an error, if there is any.
func (c *FakeMachineDeployments) Update(machineDeployment *wing.MachineDeployment) (result *wing.MachineDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(machinedeploymentsResource, c.ns, machineDeployment), &wing.MachineDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*wing.MachineDeployment), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMachineDeployments) UpdateStatus(machineDeployment *wing.MachineDeployment) (*wing.MachineDeployment, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(machinedeploymentsResource, "status", c.ns, machineDeployment), &wing.MachineDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*wing.MachineDeployment), err
}

// Delete takes name of the machineDeployment and deletes it. Returns an error if one occurs.
func (c *FakeMachineDeployments) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(machinedeploymentsResource, c.ns, name), &wing.MachineDeployment{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMachineDeployments) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(machinedeploymentsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &wing.MachineDeploymentList{})
	return err
}

// Patch applies the patch and returns the patched machineDeployment.
func (c *FakeMachineDeployments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *wing.MachineDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(machinedeploymentsResource, c.ns, name, data, subresources...), &wing.MachineDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*wing.MachineDeployment), err
}
