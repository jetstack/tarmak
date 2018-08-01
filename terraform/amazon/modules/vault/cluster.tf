resource "tarmak_vault_cluster" "vault" {
  bastion_status = "${var.bastion_status}"
  internal_fqdns = ["${aws_route53_record.per-instance.*.fqdn}"]
  vault_ca = "${element(concat(tls_self_signed_cert.ca.*.cert_pem, list("")), 0)}"
  vault_kms_key_id = "${element(split("/", var.secrets_kms_arn), 1)}"
  vault_unseal_key_name = "${data.template_file.vault_unseal_key_name.rendered}"
}