// Copyright Jetstack Ltd. See LICENSE for details.
package machinedeploymentinittime

import (
	"errors"
	"fmt"
	"io"

	"k8s.io/apiserver/pkg/admission"

	"github.com/jetstack/tarmak/pkg/apis/wing"
	informers "github.com/jetstack/tarmak/pkg/wing/client/informers/externalversions"
	listers "github.com/jetstack/tarmak/pkg/wing/client/listers/wing/v1alpha1"
)

const PluginName = "MachineDeploymentInitTime"

type machinedeploymentInitTime struct {
	*admission.Handler
	lister listers.MachineDeploymentLister
}

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return New()
	})
}

// Admit ensures that the object in-flight is of kind Flunder.
// In addition checks that the Name is not on the banned list.
// The list is stored in Fischers API objects.
func (d *machinedeploymentInitTime) Admit(a admission.Attributes) error {
	// we are only interested in machinedeployments
	if a.GetKind().GroupKind() != wing.Kind("MachineDeployment") {
		return nil
	}

	if !d.WaitForReady() {
		return admission.NewForbidden(a, fmt.Errorf("not yet ready to handle request"))
	}

	deployment, ok := a.GetObject().(*wing.MachineDeployment)
	if !ok {
		return errors.New("failed to converte obj to Machien Deployment type")
	}

	if deployment.Spec == nil {
		return errors.New("machine deployment spec cannot be nil")
	}

	if deployment.Spec.MinReplicas == nil || deployment.Spec.MaxReplicas == nil {
		return fmt.Errorf("machine deployment min or max replicas cannot be nil:  nil: min=%v max=%v",
			deployment.Spec.MinReplicas, deployment.Spec.MaxReplicas)
	}

	return nil
}

// Validate checks whether the plugin was correctly initialized.
func (d *machinedeploymentInitTime) ValidateInitialization() error {
	return nil
}

func (d *machinedeploymentInitTime) SetInternalWingInformerFactory(f informers.SharedInformerFactory) {
	d.lister = f.Wing().V1alpha1().MachineDeployments().Lister()
}

// New creates a new machines init time admission plugin
func New() (admission.Interface, error) {
	return &machinedeploymentInitTime{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}, nil
}
