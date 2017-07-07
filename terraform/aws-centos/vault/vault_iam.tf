resource "aws_iam_role" "vault" {
  name               = "${data.template_file.stack_name.rendered}-vault-${count.index+1}"
  count              = "${var.instance_count}"
  path               = "/vault-${var.environment}/"
  assume_role_policy = "${file("${path.module}/templates/role.json")}"
}

resource "aws_iam_instance_profile" "vault" {
  name  = "${data.template_file.stack_name.rendered}-vault-${count.index+1}"
  count = "${var.instance_count}"
  role  = "${element(aws_iam_role.vault.*.name, count.index)}"
}

resource "aws_iam_role_policy" "vault" {
  name   = "${data.template_file.stack_name.rendered}-vault-${count.index+1}"
  count  = "${var.instance_count}"
  role   = "${element(aws_iam_role.vault.*.name, count.index)}"
  policy = "${element(data.template_file.vault_policy.*.rendered, count.index)}"
}

data "template_file" "vault_policy" {
  template = "${file("${path.module}/templates/vault_role_policy.json")}"
  count    = "${var.instance_count}"

  vars {
    region      = "${var.region}"
    account_id  = "${data.aws_caller_identity.current.account_id}"
    volume_id   = "${element(aws_ebs_volume.vault.*.id, count.index)}"
    instance_id = "${element(aws_instance.vault.*.id, count.index)}"

    backup_bucket_prefix = "${data.terraform_remote_state.state.backups_bucket}/${data.template_file.stack_name.rendered}-vault-${count.index+1}"
    backup_bucket        = "${data.terraform_remote_state.state.backups_bucket}"

    secrets_bucket_prefix = "${data.terraform_remote_state.state.secrets_bucket}/vault-${var.environment}"
    kms_arn               = "${data.terraform_remote_state.state.secrets_kms_arn}"
    vault_unseal_key_name = "${data.template_file.vault_unseal_key_name.rendered}"
    vault_tls_cert_path           = "${element(aws_s3_bucket_object.node-certs.*.key, count.index)}"
    vault_tls_key_path            = "${element(aws_s3_bucket_object.node-keys.*.key, count.index)}"
    vault_tls_ca_path             = "${aws_s3_bucket_object.ca-cert.key}"
  }
}
