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
	v1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeConfigs implements ConfigInterface
type FakeConfigs struct {
	Fake *FakeTarmakV1alpha1
	ns   string
}

var configsResource = schema.GroupVersionResource{Group: "tarmak", Version: "v1alpha1", Resource: "configs"}

var configsKind = schema.GroupVersionKind{Group: "tarmak", Version: "v1alpha1", Kind: "Config"}

func (c *FakeConfigs) Create(config *v1alpha1.Config) (result *v1alpha1.Config, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(configsResource, c.ns, config), &v1alpha1.Config{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Config), err
}

func (c *FakeConfigs) Update(config *v1alpha1.Config) (result *v1alpha1.Config, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(configsResource, c.ns, config), &v1alpha1.Config{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Config), err
}

func (c *FakeConfigs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(configsResource, c.ns, name), &v1alpha1.Config{})

	return err
}

func (c *FakeConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(configsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ConfigList{})
	return err
}

func (c *FakeConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.Config, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(configsResource, c.ns, name), &v1alpha1.Config{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Config), err
}

func (c *FakeConfigs) List(opts v1.ListOptions) (result *v1alpha1.ConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(configsResource, configsKind, c.ns, opts), &v1alpha1.ConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ConfigList{}
	for _, item := range obj.(*v1alpha1.ConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested configs.
func (c *FakeConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(configsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched config.
func (c *FakeConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Config, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(configsResource, c.ns, name, data, subresources...), &v1alpha1.Config{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Config), err
}
