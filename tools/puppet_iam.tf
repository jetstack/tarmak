resource "aws_iam_role" "puppet_master" {
  name               = "${data.template_file.stack_name.rendered}-puppet_master"
  path               = "/"
  assume_role_policy = "${file("${path.module}/templates/role.json")}"
}

resource "aws_iam_instance_profile" "puppet_master" {
  name  = "${data.template_file.stack_name.rendered}-puppet_master"
  roles = ["${aws_iam_role.puppet_master.name}"]
}

# TODO: Warning this grants all (!!) rights to the puppet_master instance
resource "aws_iam_policy" "puppet_master" {
  name   = "${data.template_file.stack_name.rendered}-puppet_master"
  path   = "/"
  policy = "${file("${path.module}/templates/puppet_master_role_policy.json")}"
}

resource "aws_iam_policy_attachment" "puppet_master" {
  name       = "${aws_iam_policy.puppet_master.name}"
  roles      = ["${aws_iam_role.puppet_master.name}"]
  policy_arn = "${aws_iam_policy.puppet_master.arn}"
}
