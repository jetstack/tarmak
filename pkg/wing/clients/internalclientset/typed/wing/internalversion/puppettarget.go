// Copyright Jetstack Ltd. See LICENSE for details.
package internalversion

import (
	wing "github.com/jetstack/tarmak/pkg/apis/wing"
	scheme "github.com/jetstack/tarmak/pkg/wing/clients/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PuppetTargetsGetter has a method to return a PuppetTargetInterface.
// A group's client should implement this interface.
type PuppetTargetsGetter interface {
	PuppetTargets(namespace string) PuppetTargetInterface
}

// PuppetTargetInterface has methods to work with PuppetTarget resources.
type PuppetTargetInterface interface {
	Create(*wing.PuppetTarget) (*wing.PuppetTarget, error)
	Update(*wing.PuppetTarget) (*wing.PuppetTarget, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*wing.PuppetTarget, error)
	List(opts v1.ListOptions) (*wing.PuppetTargetList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *wing.PuppetTarget, err error)
	PuppetTargetExpansion
}

// puppetTargets implements PuppetTargetInterface
type puppetTargets struct {
	client rest.Interface
	ns     string
}

// newPuppetTargets returns a PuppetTargets
func newPuppetTargets(c *WingClient, namespace string) *puppetTargets {
	return &puppetTargets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the puppetTarget, and returns the corresponding puppetTarget object, and an error if there is any.
func (c *puppetTargets) Get(name string, options v1.GetOptions) (result *wing.PuppetTarget, err error) {
	result = &wing.PuppetTarget{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("puppettargets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PuppetTargets that match those selectors.
func (c *puppetTargets) List(opts v1.ListOptions) (result *wing.PuppetTargetList, err error) {
	result = &wing.PuppetTargetList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("puppettargets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested puppetTargets.
func (c *puppetTargets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("puppettargets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a puppetTarget and creates it.  Returns the server's representation of the puppetTarget, and an error, if there is any.
func (c *puppetTargets) Create(puppetTarget *wing.PuppetTarget) (result *wing.PuppetTarget, err error) {
	result = &wing.PuppetTarget{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("puppettargets").
		Body(puppetTarget).
		Do().
		Into(result)
	return
}

// Update takes the representation of a puppetTarget and updates it. Returns the server's representation of the puppetTarget, and an error, if there is any.
func (c *puppetTargets) Update(puppetTarget *wing.PuppetTarget) (result *wing.PuppetTarget, err error) {
	result = &wing.PuppetTarget{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("puppettargets").
		Name(puppetTarget.Name).
		Body(puppetTarget).
		Do().
		Into(result)
	return
}

// Delete takes name of the puppetTarget and deletes it. Returns an error if one occurs.
func (c *puppetTargets) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("puppettargets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *puppetTargets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("puppettargets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched puppetTarget.
func (c *puppetTargets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *wing.PuppetTarget, err error) {
	result = &wing.PuppetTarget{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("puppettargets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
