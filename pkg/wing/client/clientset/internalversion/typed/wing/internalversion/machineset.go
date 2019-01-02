// Copyright Jetstack Ltd. See LICENSE for details.
package internalversion

import (
	wing "github.com/jetstack/tarmak/pkg/apis/wing"
	scheme "github.com/jetstack/tarmak/pkg/wing/client/clientset/internalversion/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// MachineSetsGetter has a method to return a MachineSetInterface.
// A group's client should implement this interface.
type MachineSetsGetter interface {
	MachineSets(namespace string) MachineSetInterface
}

// MachineSetInterface has methods to work with MachineSet resources.
type MachineSetInterface interface {
	Create(*wing.MachineSet) (*wing.MachineSet, error)
	Update(*wing.MachineSet) (*wing.MachineSet, error)
	UpdateStatus(*wing.MachineSet) (*wing.MachineSet, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*wing.MachineSet, error)
	List(opts v1.ListOptions) (*wing.MachineSetList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *wing.MachineSet, err error)
	MachineSetExpansion
}

// machineSets implements MachineSetInterface
type machineSets struct {
	client rest.Interface
	ns     string
}

// newMachineSets returns a MachineSets
func newMachineSets(c *WingClient, namespace string) *machineSets {
	return &machineSets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the machineSet, and returns the corresponding machineSet object, and an error if there is any.
func (c *machineSets) Get(name string, options v1.GetOptions) (result *wing.MachineSet, err error) {
	result = &wing.MachineSet{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("machinesets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of MachineSets that match those selectors.
func (c *machineSets) List(opts v1.ListOptions) (result *wing.MachineSetList, err error) {
	result = &wing.MachineSetList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("machinesets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested machineSets.
func (c *machineSets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("machinesets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a machineSet and creates it.  Returns the server's representation of the machineSet, and an error, if there is any.
func (c *machineSets) Create(machineSet *wing.MachineSet) (result *wing.MachineSet, err error) {
	result = &wing.MachineSet{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("machinesets").
		Body(machineSet).
		Do().
		Into(result)
	return
}

// Update takes the representation of a machineSet and updates it. Returns the server's representation of the machineSet, and an error, if there is any.
func (c *machineSets) Update(machineSet *wing.MachineSet) (result *wing.MachineSet, err error) {
	result = &wing.MachineSet{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("machinesets").
		Name(machineSet.Name).
		Body(machineSet).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *machineSets) UpdateStatus(machineSet *wing.MachineSet) (result *wing.MachineSet, err error) {
	result = &wing.MachineSet{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("machinesets").
		Name(machineSet.Name).
		SubResource("status").
		Body(machineSet).
		Do().
		Into(result)
	return
}

// Delete takes name of the machineSet and deletes it. Returns an error if one occurs.
func (c *machineSets) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("machinesets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *machineSets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("machinesets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched machineSet.
func (c *machineSets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *wing.MachineSet, err error) {
	result = &wing.MachineSet{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("machinesets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
