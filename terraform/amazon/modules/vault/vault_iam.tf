resource "aws_iam_role" "vault" {
  name               = "${data.template_file.stack_name.rendered}-vault-${count.index+1}"
  count              = "${var.vault_min_instance_count}"
  path               = "/vault-${var.environment}/"
  assume_role_policy = "${file("${path.module}/templates/role.json")}"
}

resource "aws_iam_instance_profile" "vault" {
  name  = "${data.template_file.stack_name.rendered}-vault-${count.index+1}"
  count = "${var.vault_min_instance_count}"
  role  = "${element(aws_iam_role.vault.*.name, count.index)}"
}

resource "aws_iam_role_policy" "vault" {
  name   = "${data.template_file.stack_name.rendered}-vault-${count.index+1}"
  count  = "${var.vault_min_instance_count}"
  role   = "${element(aws_iam_role.vault.*.name, count.index)}"
  policy = "${element(data.template_file.vault_policy.*.rendered, count.index)}"
}

resource "aws_iam_policy" "vault_tarmak_bucket_read" {
  name   = "${data.template_file.stack_name.rendered}-vault-${count.index+1}-tarmak-bucket-read"
  count  = "${var.vault_min_instance_count}"
  policy = "${element(data.template_file.iam_vault_tarmak_bucket_read.*.rendered, count.index)}"
}

resource "aws_iam_policy_attachment" "vault_tarmak_bucket_read" {
  name       = "${data.template_file.stack_name.rendered}-vault-${count.index+1}tarmak-bucket-read"
  roles      = ["${element(aws_iam_role.vault.*.name, count.index)}"]
  count      = "${var.vault_min_instance_count}"
  policy_arn = "${element(aws_iam_policy.vault_tarmak_bucket_read.*.arn, count.index)}"
}

data "template_file" "vault_policy" {
  template = "${file("${path.module}/templates/vault_role_policy.json")}"
  count    = "${var.vault_min_instance_count}"

  vars {
    region      = "${var.region}"
    account_id  = "${data.aws_caller_identity.current.account_id}"
    volume_id   = "${element(aws_ebs_volume.vault.*.id, count.index)}"
    instance_id = "${element(aws_instance.vault.*.id, count.index)}"

    backup_bucket_prefix = "${var.backups_bucket}/${data.template_file.stack_name.rendered}-vault-${count.index+1}"
    backup_bucket        = "${var.backups_bucket}"

    secrets_bucket                = "${var.secrets_bucket}"
    vault_tls_cert_path           = "${element(aws_s3_bucket_object.node-certs.*.key, count.index)}"
    vault_tls_key_path            = "${element(aws_s3_bucket_object.node-keys.*.key, count.index)}"
    vault_tls_ca_path             = "${aws_s3_bucket_object.ca-cert.key}"
    vault_unsealer_ssm_key_prefix = "${local.vault_unseal_key_name}"
  }
}

data "template_file" "iam_vault_tarmak_bucket_read" {
  template = "${file("${path.module}/templates/iam_tarmak_bucket_read.json")}"
  count    = "${var.vault_min_instance_count}"

  vars {
    puppet_tar_gz_bucket_path    = "${var.secrets_bucket}/${aws_s3_bucket_object.latest-puppet-hash.key}"
    puppet_tar_gz_bucket_postfix = "${var.secrets_bucket}/${data.template_file.stack_name.rendered}/puppet-manifests/*-puppet.tar.gz"
    wing_binary_path             = "${var.secrets_bucket}/${data.template_file.stack_name.rendered}/wing-*"
    vault_unsealer_kms_key_id    = "${var.vault_kms_key_id}"
  }
}

resource "aws_iam_policy_attachment" "vault_additional_policies" {
  name       = "${data.template_file.stack_name.rendered}-vault-additional-policy-${count.index+1}"
  roles      = ["${aws_iam_role.vault.*.name}"]
  count      = "${length(var.vault_iam_additional_policy_arns)}"
  policy_arn = "${element(var.vault_iam_additional_policy_arns, count.index)}"
}

resource "aws_iam_role_policy_attachment" "vault_wing_binary_read" {
  role       = "${element(aws_iam_role.vault.*.name, count.index)}"
  policy_arn = "${var.wing_binary_read_policy_arn}"
  count      = "${var.vault_min_instance_count}"
}

resource "aws_iam_policy_attachment" "vault_tagging_control_lambda_invoke" {
  name       = "${data.template_file.stack_name.rendered}-tagging-control-lambda-invoke-${count.index+1}"
  roles      = ["${aws_iam_role.vault.*.name}"]
  count      = "${var.vault_min_instance_count}"
  policy_arn = "${var.tagging_control_policy_arn}"
}
