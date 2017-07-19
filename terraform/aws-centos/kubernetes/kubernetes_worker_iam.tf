resource "aws_iam_role" "kubernetes_worker" {
  name               = "kubernetes.${data.template_file.stack_name.rendered}.kubernetes_worker"
  path               = "/"
  assume_role_policy = "${file("${path.module}/templates/role.json")}"
}

resource "aws_iam_instance_profile" "kubernetes_worker" {
  name = "kubernetes.${data.template_file.stack_name.rendered}.kubernetes_worker"
  role = "${aws_iam_role.kubernetes_worker.name}"
}

data "template_file" "kubernetes_worker_iam" {
  template = "${file("${path.module}/templates/kubernetes_worker_role_policy.json")}"

  vars {
    puppet_tar_gz_bucket_path = "${data.terraform_remote_state.hub_state.secrets_bucket}/${aws_s3_bucket_object.puppet-tar-gz.key}"
  }
}

resource "aws_iam_policy" "kubernetes_worker" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.kubernetes_worker"
  path   = "/"
  policy = "${data.template_file.kubernetes_worker_iam.rendered}"
}

resource "aws_iam_policy_attachment" "kubernetes_worker" {
  name       = "${aws_iam_policy.kubernetes_worker.name}"
  roles      = ["${aws_iam_role.kubernetes_worker.name}"]
  policy_arn = "${aws_iam_policy.kubernetes_worker.arn}"
}
