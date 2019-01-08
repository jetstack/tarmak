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

package baninstance

import (
	"errors"
	"fmt"
	"io"
	"time"

	"k8s.io/apiserver/pkg/admission"

	"github.com/jetstack/tarmak/pkg/apis/wing"
	"github.com/jetstack/tarmak/pkg/wing/admission/winginitializer"
	informers "github.com/jetstack/tarmak/pkg/wing/client/informers/internalversion"
	listers "github.com/jetstack/tarmak/pkg/wing/client/listers/wing/internalversion"
)

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register("BanInstance", func(config io.Reader) (admission.Interface, error) {
		return New()
	})
}

type DisallowInstance struct {
	*admission.Handler
	lister listers.InstanceLister
}

var _ = winginitializer.WantsInternalWingInformerFactory(&DisallowInstance{})

// Admit ensures that the object in-flight is of kind Instance.
// In addition checks that the Name is not on the banned list.
// The list is stored in Instances API objects.
func (d *DisallowInstance) Admit(a admission.Attributes) error {
	// we are only interested in instancess
	if a.GetKind().GroupKind() != wing.Kind("Instance") {
		return nil
	}

	if !d.WaitForReady() {
		return admission.NewForbidden(a, fmt.Errorf("not yet ready to handle request"))
	}

	instance, ok := a.GetObject().(*wing.Instance)
	if !ok {
		return errors.New("unexpected object time")
	}

	if instance.Status != nil {
		if instance.Status.Converge != nil && instance.Status.Converge.LastUpdateTimestamp.IsZero() {
			instance.Status.Converge.LastUpdateTimestamp.Time = time.Now()
		}
		if instance.Status.DryRun != nil && instance.Status.DryRun.LastUpdateTimestamp.IsZero() {
			instance.Status.DryRun.LastUpdateTimestamp.Time = time.Now()
		}
	}
	if instance.Spec != nil {
		if instance.Spec.Converge != nil && instance.Spec.Converge.RequestTimestamp.IsZero() {
			instance.Spec.Converge.RequestTimestamp.Time = time.Now()
		}
		if instance.Spec.DryRun != nil && instance.Spec.DryRun.RequestTimestamp.IsZero() {
			instance.Spec.DryRun.RequestTimestamp.Time = time.Now()
		}
	}

	return nil
}

// SetInternalWingInformerFactory gets Lister from SharedInformerFactory.
// The lister knows how to lists Instances.
func (d *DisallowInstance) SetInternalWingInformerFactory(f informers.SharedInformerFactory) {
	d.lister = f.Wing().InternalVersion().Instances().Lister()
	d.SetReadyFunc(f.Wing().InternalVersion().Instances().Informer().HasSynced)
}

// ValidaValidateInitializationte checks whether the plugin was correctly initialized.
func (d *DisallowInstance) ValidateInitialization() error {
	if d.lister == nil {
		return fmt.Errorf("missing fischer lister")
	}
	return nil
}

// New creates a new ban flunder admission plugin
func New() (*DisallowInstance, error) {
	return &DisallowInstance{
		Handler: admission.NewHandler(admission.Create),
	}, nil
}
