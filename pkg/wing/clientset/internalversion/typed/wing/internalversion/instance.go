// Copyright Jetstack Ltd. See LICENSE for details.
package internalversion

import (
	wing "github.com/jetstack/tarmak/pkg/apis/wing"
	scheme "github.com/jetstack/tarmak/pkg/wing/clientset/internalversion/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// InstancesGetter has a method to return a InstanceInterface.
// A group's client should implement this interface.
type InstancesGetter interface {
	Instances(namespace string) InstanceInterface
}

// InstanceInterface has methods to work with Instance resources.
type InstanceInterface interface {
	Create(*wing.Instance) (*wing.Instance, error)
	Update(*wing.Instance) (*wing.Instance, error)
	UpdateStatus(*wing.Instance) (*wing.Instance, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*wing.Instance, error)
	List(opts v1.ListOptions) (*wing.InstanceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *wing.Instance, err error)
	InstanceExpansion
}

// instances implements InstanceInterface
type instances struct {
	client rest.Interface
	ns     string
}

// newInstances returns a Instances
func newInstances(c *WingClient, namespace string) *instances {
	return &instances{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the instance, and returns the corresponding instance object, and an error if there is any.
func (c *instances) Get(name string, options v1.GetOptions) (result *wing.Instance, err error) {
	result = &wing.Instance{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("instances").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Instances that match those selectors.
func (c *instances) List(opts v1.ListOptions) (result *wing.InstanceList, err error) {
	result = &wing.InstanceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("instances").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested instances.
func (c *instances) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("instances").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a instance and creates it.  Returns the server's representation of the instance, and an error, if there is any.
func (c *instances) Create(instance *wing.Instance) (result *wing.Instance, err error) {
	result = &wing.Instance{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("instances").
		Body(instance).
		Do().
		Into(result)
	return
}

// Update takes the representation of a instance and updates it. Returns the server's representation of the instance, and an error, if there is any.
func (c *instances) Update(instance *wing.Instance) (result *wing.Instance, err error) {
	result = &wing.Instance{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("instances").
		Name(instance.Name).
		Body(instance).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *instances) UpdateStatus(instance *wing.Instance) (result *wing.Instance, err error) {
	result = &wing.Instance{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("instances").
		Name(instance.Name).
		SubResource("status").
		Body(instance).
		Do().
		Into(result)
	return
}

// Delete takes name of the instance and deletes it. Returns an error if one occurs.
func (c *instances) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("instances").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *instances) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("instances").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched instance.
func (c *instances) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *wing.Instance, err error) {
	result = &wing.Instance{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("instances").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
