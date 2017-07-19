resource "aws_iam_role" "kubernetes_master" {
  name               = "kubernetes.${data.template_file.stack_name.rendered}.kubernetes_master"
  path               = "/"
  assume_role_policy = "${file("${path.module}/templates/role.json")}"
}

resource "aws_iam_instance_profile" "kubernetes_master" {
  name = "kubernetes.${data.template_file.stack_name.rendered}.kubernetes_master"
  role = "${aws_iam_role.kubernetes_master.name}"
}

data "template_file" "kubernetes_master_iam" {
  template = "${file("${path.module}/templates/kubernetes_master_role_policy.json")}"

  vars {
    puppet_tar_gz_bucket_path = "${data.terraform_remote_state.hub_state.secrets_bucket}/${aws_s3_bucket_object.puppet-tar-gz.key}"
  }
}

resource "aws_iam_policy" "kubernetes_master" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.kubernetes_master"
  path   = "/"
  policy = "${data.template_file.kubernetes_master_iam.rendered}"
}

resource "aws_iam_policy_attachment" "kubernetes_master" {
  name       = "${aws_iam_policy.kubernetes_master.name}"
  roles      = ["${aws_iam_role.kubernetes_master.name}"]
  policy_arn = "${aws_iam_policy.kubernetes_master.arn}"
}
