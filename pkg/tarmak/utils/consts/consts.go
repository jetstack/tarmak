// Copyright Jetstack Ltd. See LICENSE for details.
package consts

const (
	AmazonRateLimitErr = "RequestLimitExceeded"

	DefaultPlanLocationPlaceholder = "${TARMAK_CONFIG}/${CURRENT_CLUSTER}/terraform/tarmak.plan"
	DefaultLogsPathPlaceholder     = "${TARMAK_CONFIG}/${CURRENT_CLUSTER}/${INSTANCE_POOL}.tar.gz"
	TerraformPlanFile              = "tarmak.plan"

	DefaultKubeconfigPath = "${TARMAK_CONFIG}/${CURRENT_CLUSTER}/kubeconfig"
	KubeconfigFlagName    = "public-api-endpoint"
)
