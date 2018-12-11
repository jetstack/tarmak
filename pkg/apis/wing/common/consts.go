// Copyright Jetstack Ltd. See LICENSE for details.
package common

type MachineStatusError string
type MachineManifestState string

type MachineDeploymentStrategyType string

const (
	MachineManifestStateConverging = MachineManifestState("converging")
	MachineManifestStateConverged  = MachineManifestState("converged")
	MachineManifestStateError      = MachineManifestState("error")

	// Replace the old MachineSet by new one using rolling update
	// i.e gradually scale down the old MachineSet and scale up the new one.
	RollingUpdateMachineDeploymentStrategyType MachineDeploymentStrategyType = "RollingUpdate"
)
