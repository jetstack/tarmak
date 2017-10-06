// Copyright Jetstack Ltd. See LICENSE for details.
package v1alpha1

type InstanceManifestState string

const (
	InstanceManifestStateConverging = InstanceManifestState("converging")
	InstanceManifestStateConverged  = InstanceManifestState("converged")
	InstanceManifestStateError      = InstanceManifestState("error")
)
