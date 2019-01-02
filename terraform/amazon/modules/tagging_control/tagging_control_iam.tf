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

resource "aws_security_group" "tagging_control" {
  name   = "${data.template_file.stack_name.rendered}-tagging_control"
  vpc_id = "${var.vpc_id}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-tagging_control"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_security_group_rule" "tagging_control_out_allow_all" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.tagging_control.id}"
}

resource "aws_security_group_rule" "tagging_control_in_allow_all" {
  type        = "ingress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.tagging_control.id}"
}

resource "aws_iam_policy" "tagging_control_lambda_invoke" {
  name = "${data.template_file.stack_name.rendered}.tagging_control_lambda_invoke"
  path = "/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "lambda:InvokeFunction",
      "Resource": "${aws_lambda_function.tagging_control.arn}"
    }
  ]
}
EOF
}
