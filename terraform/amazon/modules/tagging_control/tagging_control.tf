resource "aws_lambda_function" "tagging_control" {
  filename         = "tagging_control.zip"
  function_name    = "tarmak_tagging_control"
  role             = "${aws_iam_role.tagging_control.arn}"
  handler          = "tagging_control_linux_amd64"
  source_code_hash = "${base64sha256(file("tagging_control.zip"))}"
  runtime          = "go1.x"
  timeout          = "10"

  vpc_config {
    subnet_ids = ["${var.private_subnet_ids}"]
    security_group_ids = ["${aws_security_group.tagging_control.id}"]
  }

  tags {
    Name        = "${data.template_file.stack_name.rendered}-tagging_control"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }

  depends_on = ["aws_iam_role_policy.tagging_control"]
}
