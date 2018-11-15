// Copyright Jetstack Ltd. See LICENSE for details.
package consts

const (
	AmazonRateLimitErr = "RequestLimitExceeded"

	DefaultPlanLocationPlaceholder = "${TARMAK_CONFIG}/${CURRENT_CLUSTER}/terraform/tarmak.plan"
	TerraformPlanFile              = "tarmak.plan"

	DefaultKubeconfigPath = "${TARMAK_CONFIG}/${CURRENT_CLUSTER}/kubeconfig"
)
