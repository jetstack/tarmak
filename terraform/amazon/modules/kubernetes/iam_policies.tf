data "template_file" "iam_tarmak_bucket_read" {
  template = "${file("${path.module}/templates/iam_tarmak_bucket_read.json")}"

  vars {
    puppet_tar_gz_bucket_path    = "${var.secrets_bucket}/${aws_s3_bucket_object.latest-puppet-hash.key}"
    puppet_tar_gz_bucket_postfix = "${var.secrets_bucket}/${data.template_file.stack_name.rendered}/puppet-manifests/*-puppet.tar.gz"

    legacy_puppet_tar_gz_bucket_path = "${var.secrets_bucket}/${data.template_file.stack_name.rendered}/puppet.tar.gz"

    vault_kms_key_id = "${var.vault_kms_key_id}"
  }
}

resource "aws_iam_policy" "tarmak_bucket_read" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.tarmak_bucket_read"
  path   = "/"
  policy = "${data.template_file.iam_tarmak_bucket_read.rendered}"
}

resource "aws_iam_policy" "ec2_controller" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.ec2_controller"
  path   = "/"
  policy = "${file("${path.module}/templates/iam_ec2_controller.json")}"
}

resource "aws_iam_policy" "ec2_read" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.ec2_read"
  path   = "/"
  policy = "${file("${path.module}/templates/iam_ec2_read.json")}"
}

resource "aws_iam_policy" "ec2_modify_instance_attribute" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.ec2_modify_instance_attribute"
  path   = "/"
  policy = "${file("${path.module}/templates/iam_ec2_modify_instance_attribute.json")}"
}

resource "aws_iam_policy" "ecr_read" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.ecr_read"
  path   = "/"
  policy = "${file("${path.module}/templates/iam_ecr_read.json")}"
}

resource "aws_iam_policy" "elb_controller" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.elb_controller"
  path   = "/"
  policy = "${file("${path.module}/templates/iam_elb_controller.json")}"
}

resource "aws_iam_policy" "cluster_autoscaler" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.cluster_autoscaler"
  path   = "/"
  policy = "${file("${path.module}/templates/iam_cluster_autoscaler.json")}"
}
