// Copyright Jetstack Ltd. See LICENSE for details.
package instaceinittime

import (
	"errors"
	"io"
	"time"

	"k8s.io/apiserver/pkg/admission"

	"github.com/jetstack/tarmak/pkg/apis/wing"
)

const PluginName = "MachineInitTime"

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return New()
	})
}

type machineInitTime struct {
	*admission.Handler
}

// Admit ensures that the object in-flight is of kind Flunder.
// In addition checks that the Name is not on the banned list.
// The list is stored in Fischers API objects.
func (d *machineInitTime) Admit(a admission.Attributes) error {
	// we are only interested in machines
	if a.GetKind().GroupKind() != wing.Kind("Machine") {
		return nil
	}

	machine, ok := a.GetObject().(*wing.Machine)
	if !ok {
		return errors.New("unexpected object time")
	}

	if machine.Status != nil {
		if machine.Status.Converge != nil && machine.Status.Converge.LastUpdateTimestamp.IsZero() {
			machine.Status.Converge.LastUpdateTimestamp.Time = time.Now()
		}
		if machine.Status.DryRun != nil && machine.Status.DryRun.LastUpdateTimestamp.IsZero() {
			machine.Status.DryRun.LastUpdateTimestamp.Time = time.Now()
		}
	}
	if machine.Spec != nil {
		if machine.Spec.Converge != nil && machine.Spec.Converge.RequestTimestamp.IsZero() {
			machine.Spec.Converge.RequestTimestamp.Time = time.Now()
		}
		if machine.Spec.DryRun != nil && machine.Spec.DryRun.RequestTimestamp.IsZero() {
			machine.Spec.DryRun.RequestTimestamp.Time = time.Now()
		}
	}

	return nil
}

// Validate checks whether the plugin was correctly initialized.
func (d *machineInitTime) Validate() error {
	return nil
}

// New creates a new machines init time admission plugin
func New() (admission.Interface, error) {
	return &machineInitTime{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}, nil
}
