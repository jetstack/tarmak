// Copyright Jetstack Ltd. See LICENSE for details.
package v1alpha1

import (
	"time"
)

const (
	StackNameState      = "state"
	StackNameNetwork    = "network"
	StackNameTools      = "tools"
	StackNameVault      = "vault"
	StackNameKubernetes = "kubernetes"
)

const (
	ImageTagEnvironment   = "tarmak_environment"
	ImageTagBaseImageName = "tarmak_base_image_name"
)

const (
	EnvironmentTypeEmpty  = "empty"  // an environment that contains no cluster at all
	EnvironmentTypeMulti  = "multi"  // an environment that contains a hub and 0-n clusters
	EnvironmentTypeSingle = "single" // an environment that contains exactly one cluster
)

var KubernetesEpoch time.Time = time.Unix(1437436800, 0)
