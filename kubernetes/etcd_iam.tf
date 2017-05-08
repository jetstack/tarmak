resource "aws_iam_role" "etcd" {
  name               = "kubernetes.${data.template_file.stack_name.rendered}.etcd"
  path               = "/"
  assume_role_policy = "${file("${path.module}/templates/role.json")}"
}

resource "aws_iam_instance_profile" "etcd" {
  name = "kubernetes.${data.template_file.stack_name.rendered}.etcd"
  role = "${aws_iam_role.etcd.name}"
}

resource "aws_iam_policy" "etcd" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.etcd"
  path   = "/"
  policy = "${file("${path.module}/templates/etcd_role_policy.json")}"
}

resource "aws_iam_policy_attachment" "etcd" {
  name       = "${aws_iam_policy.etcd.name}"
  roles      = ["${aws_iam_role.etcd.name}"]
  policy_arn = "${aws_iam_policy.etcd.arn}"
}
