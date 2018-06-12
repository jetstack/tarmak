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

// FakePuppetTargets implements PuppetTargetInterface
type FakePuppetTargets struct {
	Fake *FakeWing
	ns   string
}

var puppettargetsResource = schema.GroupVersionResource{Group: "wing.tarmak.io", Version: "", Resource: "puppettargets"}

var puppettargetsKind = schema.GroupVersionKind{Group: "wing.tarmak.io", Version: "", Kind: "PuppetTarget"}

// Get takes name of the puppetTarget, and returns the corresponding puppetTarget object, and an error if there is any.
func (c *FakePuppetTargets) Get(name string, options v1.GetOptions) (result *wing.PuppetTarget, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(puppettargetsResource, c.ns, name), &wing.PuppetTarget{})

	if obj == nil {
		return nil, err
	}
	return obj.(*wing.PuppetTarget), err
}

// List takes label and field selectors, and returns the list of PuppetTargets that match those selectors.
func (c *FakePuppetTargets) List(opts v1.ListOptions) (result *wing.PuppetTargetList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(puppettargetsResource, puppettargetsKind, c.ns, opts), &wing.PuppetTargetList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &wing.PuppetTargetList{}
	for _, item := range obj.(*wing.PuppetTargetList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested puppetTargets.
func (c *FakePuppetTargets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(puppettargetsResource, c.ns, opts))

}

// Create takes the representation of a puppetTarget and creates it.  Returns the server's representation of the puppetTarget, and an error, if there is any.
func (c *FakePuppetTargets) Create(puppetTarget *wing.PuppetTarget) (result *wing.PuppetTarget, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(puppettargetsResource, c.ns, puppetTarget), &wing.PuppetTarget{})

	if obj == nil {
		return nil, err
	}
	return obj.(*wing.PuppetTarget), err
}

// Update takes the representation of a puppetTarget and updates it. Returns the server's representation of the puppetTarget, and an error, if there is any.
func (c *FakePuppetTargets) Update(puppetTarget *wing.PuppetTarget) (result *wing.PuppetTarget, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(puppettargetsResource, c.ns, puppetTarget), &wing.PuppetTarget{})

	if obj == nil {
		return nil, err
	}
	return obj.(*wing.PuppetTarget), err
}

// Delete takes name of the puppetTarget and deletes it. Returns an error if one occurs.
func (c *FakePuppetTargets) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(puppettargetsResource, c.ns, name), &wing.PuppetTarget{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePuppetTargets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(puppettargetsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &wing.PuppetTargetList{})
	return err
}

// Patch applies the patch and returns the patched puppetTarget.
func (c *FakePuppetTargets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *wing.PuppetTarget, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(puppettargetsResource, c.ns, name, data, subresources...), &wing.PuppetTarget{})

	if obj == nil {
		return nil, err
	}
	return obj.(*wing.PuppetTarget), err
}
