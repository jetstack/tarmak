resource "aws_iam_role" "etcd" {
  name               = "kubernetes.${data.template_file.stack_name.rendered}.etcd"
  path               = "/"
  assume_role_policy = "${file("${path.module}/templates/role.json")}"
}

resource "aws_iam_instance_profile" "etcd" {
  name = "kubernetes.${data.template_file.stack_name.rendered}.etcd"
  role = "${aws_iam_role.etcd.name}"
}

data "template_file" "etcd_iam" {
  template = "${file("${path.module}/templates/etcd_role_policy.json")}"

  vars {
    puppet_tar_gz_bucket_path = "${data.terraform_remote_state.hub_state.secrets_bucket}/${aws_s3_bucket_object.puppet-tar-gz.key}"
  }
}

resource "aws_iam_policy" "etcd" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.etcd"
  path   = "/"
  policy = "${data.template_file.etcd_iam.rendered}"
}

resource "aws_iam_policy_attachment" "etcd" {
  name       = "${aws_iam_policy.etcd.name}"
  roles      = ["${aws_iam_role.etcd.name}"]
  policy_arn = "${aws_iam_policy.etcd.arn}"
}
