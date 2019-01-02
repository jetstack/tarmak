// Copyright Jetstack Ltd. See LICENSE for details.

// This file was automatically generated by informer-gen

package wing

import (
	internalinterfaces "github.com/jetstack/tarmak/pkg/wing/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// Machines returns a MachineInformer.
	Machines() MachineInformer
	// MachineDeployments returns a MachineDeploymentInformer.
	MachineDeployments() MachineDeploymentInformer
	// MachineSets returns a MachineSetInformer.
	MachineSets() MachineSetInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// Machines returns a MachineInformer.
func (v *version) Machines() MachineInformer {
	return &machineInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MachineDeployments returns a MachineDeploymentInformer.
func (v *version) MachineDeployments() MachineDeploymentInformer {
	return &machineDeploymentInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MachineSets returns a MachineSetInformer.
func (v *version) MachineSets() MachineSetInformer {
	return &machineSetInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
