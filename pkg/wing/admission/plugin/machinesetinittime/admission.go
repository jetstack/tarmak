// Copyright Jetstack Ltd. See LICENSE for details.
package machinesetinittime

import (
	"errors"
	"fmt"
	"io"

	"k8s.io/apiserver/pkg/admission"

	"github.com/jetstack/tarmak/pkg/apis/wing"
)

const PluginName = "MachineSetInitTime"

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return New()
	})
}

type machinesetInitTime struct {
	*admission.Handler
}

// Admit ensures that the object in-flight is of kind Flunder.
// In addition checks that the Name is not on the banned list.
// The list is stored in Fischers API objects.
func (d *machinesetInitTime) Admit(a admission.Attributes) error {
	// we are only interested in machinesetss
	if a.GetKind().GroupKind() != wing.Kind("MachineSet") {
		return nil
	}

	machineset, ok := a.GetObject().(*wing.MachineSet)
	if !ok {
		return errors.New("unexpected object type")
	}

	if machineset.Spec == nil {
		return errors.New("machineset spec cannot be nil")
	}

	if machineset.Spec.MinReplicas == nil || machineset.Spec.MaxReplicas == nil {
		return fmt.Errorf("machineset min or max replicas cannot be nil: min=%v max=%v",
			machineset.Spec.MinReplicas, machineset.Spec.MaxReplicas)
	}

	return nil
}

// Validate checks whether the plugin was correctly initialized.
func (d *machinesetInitTime) Validate() error {
	return nil
}

// New creates a new machines init time admission plugin
func New() (admission.Interface, error) {
	return &machinesetInitTime{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}, nil
}
