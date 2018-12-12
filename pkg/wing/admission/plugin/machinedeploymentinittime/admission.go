// Copyright Jetstack Ltd. See LICENSE for details.
package machinedeploymentinittime

import (
	"errors"
	"fmt"
	"io"

	"k8s.io/apiserver/pkg/admission"

	"github.com/jetstack/tarmak/pkg/apis/wing"
)

const PluginName = "MachineDeploymentInitTime"

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return New()
	})
}

type machinedeploymentInitTime struct {
	*admission.Handler
}

// Admit ensures that the object in-flight is of kind Flunder.
// In addition checks that the Name is not on the banned list.
// The list is stored in Fischers API objects.
func (d *machinedeploymentInitTime) Admit(a admission.Attributes) error {
	// we are only interested in machinedeploymentss
	if a.GetKind().GroupKind() != wing.Kind("MachineDeployment") {
		return nil
	}

	machinedeployment, ok := a.GetObject().(*wing.MachineDeployment)
	if !ok {
		return errors.New("unexpected object time")
	}

	if machinedeployment.Spec == nil {
		return fmt.Errorf("expected machinedeployment replicas to be set, got=%v", machinedeployment.Spec)
	}

	if machinedeployment.Spec.MinReplicas == nil {
		return fmt.Errorf("expected machinedeployment min replicas to be set, got=%v", machinedeployment.Spec.MinReplicas)
	}
	if machinedeployment.Spec.MaxReplicas == nil {
		return fmt.Errorf("expected machinedeployment max replicas to be set, got=%v", machinedeployment.Spec.MaxReplicas)
	}

	return nil
}

// Validate checks whether the plugin was correctly initialized.
func (d *machinedeploymentInitTime) Validate() error {
	return nil
}

// New creates a new machines init time admission plugin
func New() (admission.Interface, error) {
	return &machinedeploymentInitTime{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}, nil
}
