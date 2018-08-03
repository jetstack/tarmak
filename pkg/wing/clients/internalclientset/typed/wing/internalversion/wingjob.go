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

// WingJobsGetter has a method to return a WingJobInterface.
// A group's client should implement this interface.
type WingJobsGetter interface {
	WingJobs(namespace string) WingJobInterface
}

// WingJobInterface has methods to work with WingJob resources.
type WingJobInterface interface {
	Create(*wing.WingJob) (*wing.WingJob, error)
	Update(*wing.WingJob) (*wing.WingJob, error)
	UpdateStatus(*wing.WingJob) (*wing.WingJob, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*wing.WingJob, error)
	List(opts v1.ListOptions) (*wing.WingJobList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *wing.WingJob, err error)
	WingJobExpansion
}

// wingJobs implements WingJobInterface
type wingJobs struct {
	client rest.Interface
	ns     string
}

// newWingJobs returns a WingJobs
func newWingJobs(c *WingClient, namespace string) *wingJobs {
	return &wingJobs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the wingJob, and returns the corresponding wingJob object, and an error if there is any.
func (c *wingJobs) Get(name string, options v1.GetOptions) (result *wing.WingJob, err error) {
	result = &wing.WingJob{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("wingjobs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of WingJobs that match those selectors.
func (c *wingJobs) List(opts v1.ListOptions) (result *wing.WingJobList, err error) {
	result = &wing.WingJobList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("wingjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested wingJobs.
func (c *wingJobs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("wingjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a wingJob and creates it.  Returns the server's representation of the wingJob, and an error, if there is any.
func (c *wingJobs) Create(wingJob *wing.WingJob) (result *wing.WingJob, err error) {
	result = &wing.WingJob{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("wingjobs").
		Body(wingJob).
		Do().
		Into(result)
	return
}

// Update takes the representation of a wingJob and updates it. Returns the server's representation of the wingJob, and an error, if there is any.
func (c *wingJobs) Update(wingJob *wing.WingJob) (result *wing.WingJob, err error) {
	result = &wing.WingJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("wingjobs").
		Name(wingJob.Name).
		Body(wingJob).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *wingJobs) UpdateStatus(wingJob *wing.WingJob) (result *wing.WingJob, err error) {
	result = &wing.WingJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("wingjobs").
		Name(wingJob.Name).
		SubResource("status").
		Body(wingJob).
		Do().
		Into(result)
	return
}

// Delete takes name of the wingJob and deletes it. Returns an error if one occurs.
func (c *wingJobs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("wingjobs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *wingJobs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("wingjobs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched wingJob.
func (c *wingJobs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *wing.WingJob, err error) {
	result = &wing.WingJob{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("wingjobs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
