output "vault_ca" {
  value = "${element(concat(tls_self_signed_cert.ca.*.cert_pem, list("")), 0)}"
}

output "vault_url" {
  value = "https://${element(concat(aws_route53_record.endpoint.*.fqdn, list("")), 0)}:8200"
}

output "vault_kms_key_id" {
  value = "${var.secrets_kms_arn}"
}

output "vault_unseal_key_name" {
  value = "${data.template_file.vault_unseal_key_name.rendered}"
}

output "instance_fqdns" {
  value = ["${aws_route53_record.per-instance.*.fqdn}"]
}

output "vault_security_group_id" {
  value = "${element(concat(aws_security_group.vault.*.id, list("")), 0)}"
}

output "vault_aws_caller_identity_current_account_id" {
  value = "${data.aws_caller_identity.current.account_id}"
}
