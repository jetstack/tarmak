// Copyright Jetstack Ltd. See LICENSE for details.
package fake

import (
	v1alpha1 "github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeWingJobs implements WingJobInterface
type FakeWingJobs struct {
	Fake *FakeWingV1alpha1
	ns   string
}

var wingjobsResource = schema.GroupVersionResource{Group: "wing.tarmak.io", Version: "v1alpha1", Resource: "wingjobs"}

var wingjobsKind = schema.GroupVersionKind{Group: "wing.tarmak.io", Version: "v1alpha1", Kind: "WingJob"}

// Get takes name of the wingJob, and returns the corresponding wingJob object, and an error if there is any.
func (c *FakeWingJobs) Get(name string, options v1.GetOptions) (result *v1alpha1.WingJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(wingjobsResource, c.ns, name), &v1alpha1.WingJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WingJob), err
}

// List takes label and field selectors, and returns the list of WingJobs that match those selectors.
func (c *FakeWingJobs) List(opts v1.ListOptions) (result *v1alpha1.WingJobList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(wingjobsResource, wingjobsKind, c.ns, opts), &v1alpha1.WingJobList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.WingJobList{}
	for _, item := range obj.(*v1alpha1.WingJobList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested wingJobs.
func (c *FakeWingJobs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(wingjobsResource, c.ns, opts))

}

// Create takes the representation of a wingJob and creates it.  Returns the server's representation of the wingJob, and an error, if there is any.
func (c *FakeWingJobs) Create(wingJob *v1alpha1.WingJob) (result *v1alpha1.WingJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(wingjobsResource, c.ns, wingJob), &v1alpha1.WingJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WingJob), err
}

// Update takes the representation of a wingJob and updates it. Returns the server's representation of the wingJob, and an error, if there is any.
func (c *FakeWingJobs) Update(wingJob *v1alpha1.WingJob) (result *v1alpha1.WingJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(wingjobsResource, c.ns, wingJob), &v1alpha1.WingJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WingJob), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeWingJobs) UpdateStatus(wingJob *v1alpha1.WingJob) (*v1alpha1.WingJob, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(wingjobsResource, "status", c.ns, wingJob), &v1alpha1.WingJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WingJob), err
}

// Delete takes name of the wingJob and deletes it. Returns an error if one occurs.
func (c *FakeWingJobs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(wingjobsResource, c.ns, name), &v1alpha1.WingJob{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeWingJobs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(wingjobsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.WingJobList{})
	return err
}

// Patch applies the patch and returns the patched wingJob.
func (c *FakeWingJobs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.WingJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(wingjobsResource, c.ns, name, data, subresources...), &v1alpha1.WingJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WingJob), err
}
