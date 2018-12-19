resource "aws_iam_role" "tagging_control" {
  name               = "${data.template_file.stack_name.rendered}-tagging_control"
  path               = "/tagging_control-${var.environment}/"
  assume_role_policy = "${file("${path.module}/templates/role.json")}"
}

resource "aws_iam_role_policy" "tagging_control" {
  name   = "${data.template_file.stack_name.rendered}-tagging_control"
  role   = "${aws_iam_role.tagging_control.name}"
  policy = "${data.template_file.tagging_control_policy.rendered}"
}

data "template_file" "tagging_control_policy" {
  template = "${file("${path.module}/templates/tagging_control_policy.json")}"
}

resource "aws_lambda_function" "tagging_control" {
  filename         = "tagging_control.zip"
  function_name    = "tarmak_tagging_control"
  role             = "${aws_iam_role.tagging_control.arn}"
  handler          = "tagging_control_linux_amd64"
  source_code_hash = "${base64sha256(file("tagging_control.zip"))}"
  runtime          = "go1.x"

  vpc_config {
    subnet_ids = ["${var.public_subnet_ids[0]}"]
    security_group_ids = ["${var.bastion_security_group_id}"]
  }

  tags {
    Name        = "${data.template_file.stack_name.rendered}-tagging_control"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }

  depends_on = ["aws_iam_role_policy.tagging_control"]
}
