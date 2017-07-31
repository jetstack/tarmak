/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fake

import (
	tarmak "github.com/jetstack/tarmak/pkg/apis/tarmak"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeEnvironments implements EnvironmentInterface
type FakeEnvironments struct {
	Fake *FakeTarmak
	ns   string
}

var environmentsResource = schema.GroupVersionResource{Group: "tarmak", Version: "", Resource: "environments"}

var environmentsKind = schema.GroupVersionKind{Group: "tarmak", Version: "", Kind: "Environment"}

func (c *FakeEnvironments) Create(environment *tarmak.Environment) (result *tarmak.Environment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(environmentsResource, c.ns, environment), &tarmak.Environment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarmak.Environment), err
}

func (c *FakeEnvironments) Update(environment *tarmak.Environment) (result *tarmak.Environment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(environmentsResource, c.ns, environment), &tarmak.Environment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarmak.Environment), err
}

func (c *FakeEnvironments) UpdateStatus(environment *tarmak.Environment) (*tarmak.Environment, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(environmentsResource, "status", c.ns, environment), &tarmak.Environment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarmak.Environment), err
}

func (c *FakeEnvironments) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(environmentsResource, c.ns, name), &tarmak.Environment{})

	return err
}

func (c *FakeEnvironments) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(environmentsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &tarmak.EnvironmentList{})
	return err
}

func (c *FakeEnvironments) Get(name string, options v1.GetOptions) (result *tarmak.Environment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(environmentsResource, c.ns, name), &tarmak.Environment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarmak.Environment), err
}

func (c *FakeEnvironments) List(opts v1.ListOptions) (result *tarmak.EnvironmentList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(environmentsResource, environmentsKind, c.ns, opts), &tarmak.EnvironmentList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &tarmak.EnvironmentList{}
	for _, item := range obj.(*tarmak.EnvironmentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested environments.
func (c *FakeEnvironments) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(environmentsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched environment.
func (c *FakeEnvironments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *tarmak.Environment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(environmentsResource, c.ns, name, data, subresources...), &tarmak.Environment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarmak.Environment), err
}
