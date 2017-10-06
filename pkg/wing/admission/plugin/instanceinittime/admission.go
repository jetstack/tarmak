// Copyright Jetstack Ltd. See LICENSE for details.
package instaceinittime

import (
	"errors"
	"io"
	"time"

	"k8s.io/apiserver/pkg/admission"

	"github.com/jetstack/tarmak/pkg/apis/wing"
)

const PluginName = "InstanceInitTime"

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return New()
	})
}

type instanceInitTime struct {
	*admission.Handler
}

// Admit ensures that the object in-flight is of kind Flunder.
// In addition checks that the Name is not on the banned list.
// The list is stored in Fischers API objects.
func (d *instanceInitTime) Admit(a admission.Attributes) error {
	// we are only interested in instances
	if a.GetKind().GroupKind() != wing.Kind("Instance") {
		return nil
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

// Validate checks whether the plugin was correctly initialized.
func (d *instanceInitTime) Validate() error {
	return nil
}

// New creates a new instances init time admission plugin
func New() (admission.Interface, error) {
	return &instanceInitTime{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}, nil
}
