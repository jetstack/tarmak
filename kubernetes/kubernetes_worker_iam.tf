resource "aws_iam_role" "kubernetes_worker" {
  name               = "kubernetes.${data.template_file.stack_name.rendered}.kubernetes_worker"
  path               = "/"
  assume_role_policy = "${file("${path.module}/templates/role.json")}"
}

resource "aws_iam_instance_profile" "kubernetes_worker" {
  name  = "kubernetes.${data.template_file.stack_name.rendered}.kubernetes_worker"
  roles = ["${aws_iam_role.kubernetes_worker.name}"]
}

resource "aws_iam_policy" "kubernetes_worker" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.kubernetes_worker"
  path   = "/"
  policy = "${file("${path.module}/templates/kubernetes_worker_role_policy.json")}"
}

resource "aws_iam_policy_attachment" "kubernetes_worker" {
  name       = "${aws_iam_policy.kubernetes_worker.name}"
  roles      = ["${aws_iam_role.kubernetes_worker.name}"]
  policy_arn = "${aws_iam_policy.kubernetes_worker.arn}"
}
