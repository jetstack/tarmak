resource "aws_iam_role" "kubernetes_master" {
  name               = "kubernetes.${data.template_file.stack_name.rendered}.kubernetes_master"
  path               = "/"
  assume_role_policy = "${file("${path.module}/templates/role.json")}"
}

resource "aws_iam_instance_profile" "kubernetes_master" {
  name  = "kubernetes.${data.template_file.stack_name.rendered}.kubernetes_master"
  roles = ["${aws_iam_role.kubernetes_master.name}"]
}

resource "aws_iam_policy" "kubernetes_master" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.kubernetes_master"
  path   = "/"
  policy = "${file("${path.module}/templates/kubernetes_master_role_policy.json")}"
}

resource "aws_iam_policy_attachment" "kubernetes_master" {
  name       = "${aws_iam_policy.kubernetes_master.name}"
  roles      = ["${aws_iam_role.kubernetes_master.name}"]
  policy_arn = "${aws_iam_policy.kubernetes_master.arn}"
}
