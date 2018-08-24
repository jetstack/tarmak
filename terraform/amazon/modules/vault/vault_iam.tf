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
    vault_unsealer_kms_key_id     = "${var.secrets_kms_arn}"
    vault_unsealer_ssm_key_prefix = "${data.template_file.vault_unseal_key_name.rendered}"

    puppet_tar_gz_bucket_path = "${var.secrets_bucket}/${aws_s3_bucket_object.puppet-tar-gz.key}"
  }
}
