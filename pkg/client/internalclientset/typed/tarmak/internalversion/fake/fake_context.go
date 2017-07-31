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

// FakeContexts implements ContextInterface
type FakeContexts struct {
	Fake *FakeTarmak
	ns   string
}

var contextsResource = schema.GroupVersionResource{Group: "tarmak", Version: "", Resource: "contexts"}

var contextsKind = schema.GroupVersionKind{Group: "tarmak", Version: "", Kind: "Context"}

func (c *FakeContexts) Create(context *tarmak.Context) (result *tarmak.Context, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(contextsResource, c.ns, context), &tarmak.Context{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarmak.Context), err
}

func (c *FakeContexts) Update(context *tarmak.Context) (result *tarmak.Context, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(contextsResource, c.ns, context), &tarmak.Context{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarmak.Context), err
}

func (c *FakeContexts) UpdateStatus(context *tarmak.Context) (*tarmak.Context, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(contextsResource, "status", c.ns, context), &tarmak.Context{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarmak.Context), err
}

func (c *FakeContexts) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(contextsResource, c.ns, name), &tarmak.Context{})

	return err
}

func (c *FakeContexts) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(contextsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &tarmak.ContextList{})
	return err
}

func (c *FakeContexts) Get(name string, options v1.GetOptions) (result *tarmak.Context, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(contextsResource, c.ns, name), &tarmak.Context{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarmak.Context), err
}

func (c *FakeContexts) List(opts v1.ListOptions) (result *tarmak.ContextList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(contextsResource, contextsKind, c.ns, opts), &tarmak.ContextList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &tarmak.ContextList{}
	for _, item := range obj.(*tarmak.ContextList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested contexts.
func (c *FakeContexts) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(contextsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched context.
func (c *FakeContexts) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *tarmak.Context, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(contextsResource, c.ns, name, data, subresources...), &tarmak.Context{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarmak.Context), err
}
