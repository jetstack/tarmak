data "template_file" "iam_tarmak_bucket_read" {
  template = "${file("${path.module}/templates/iam_tarmak_bucket_read.json")}"

  vars {
    puppet_tar_gz_bucket_path = "${data.terraform_remote_state.hub_state.secrets_bucket}/${aws_s3_bucket_object.puppet-tar-gz.key}"
  }
}

resource "aws_iam_policy" "tarmak_bucket_read" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.tarmak_bucket_read"
  path   = "/"
  policy = "${data.template_file.iam_tarmak_bucket_read.rendered}"
}

resource "aws_iam_policy" "ec2_full" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.ec2_full"
  path   = "/"
  policy = "${file("${path.module}/templates/iam_ec2_full.json")}"
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

resource "aws_iam_policy" "elb_full" {
  name   = "kubernetes.${data.template_file.stack_name.rendered}.elb_full"
  path   = "/"
  policy = "${file("${path.module}/templates/iam_elb_full.json")}"
}
