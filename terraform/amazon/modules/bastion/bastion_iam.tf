resource "aws_iam_role" "bastion" {
  name               = "${data.template_file.stack_name.rendered}-bastion"
  count              = "${var.bastion_min_instance_count}"
  path               = "/bastion-${var.environment}/"
  assume_role_policy = "${file("${path.module}/templates/role.json")}"
}

resource "aws_iam_instance_profile" "bastion" {
  name  = "${data.template_file.stack_name.rendered}-bastion"
  count = "${var.bastion_min_instance_count}"
  role  = "${aws_iam_role.bastion.name}"
}

resource "aws_iam_policy_attachment" "bastion_additional_policy" {
  name       = "${data.template_file.stack_name.rendered}-bastion-additional-policy-${count.index+1}"
  roles      = ["${aws_iam_role.bastion.name}"]
  count      = "${length(var.bastion_iam_additional_policy_arns)}"
  policy_arn = "${element(var.bastion_iam_additional_policy_arns, count.index)}"
}

data "template_file" "wing_binary_read" {
  template = "${file("${path.module}/templates/wing_binary_read.json")}"

  vars {
    wing_binary_path = "${var.secrets_bucket}/${data.template_file.stack_name.rendered}/wing-*"
  }
}

resource "aws_iam_policy" "wing_binary_read" {
  name   = "bastion.${data.template_file.stack_name.rendered}.wing_binary_read"
  path   = "/"
  policy = "${data.template_file.wing_binary_read.rendered}"
}

resource "aws_iam_policy_attachment" "wing_binary_read" {
  name       = "${data.template_file.stack_name.rendered}-wing-binary"
  roles      = ["${aws_iam_role.bastion.name}"]
  policy_arn = "${aws_iam_policy.wing_binary_read.arn}"
}
