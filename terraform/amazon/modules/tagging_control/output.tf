output "tagging_control_policy_arn" {
  value = "${aws_iam_policy.tagging_control_lambda_invoke.arn}"
}
