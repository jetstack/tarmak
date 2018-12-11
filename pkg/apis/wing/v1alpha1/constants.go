// Copyright Jetstack Ltd. See LICENSE for details.
package v1alpha1

type MachineManifestState string

const (
	MachineManifestStateConverging = MachineManifestState("converging")
	MachineManifestStateConverged  = MachineManifestState("converged")
	MachineManifestStateError      = MachineManifestState("error")
)
