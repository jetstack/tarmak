output "vault_ca" {
  value = "${tls_self_signed_cert.ca.cert_pem}"
}

output "vault_url" {
  value = "https://${aws_route53_record.endpoint.fqdn}:8200"
}

output "vault_kms_key_id" {
  value = "${element(split("/", var.secrets_kms_arn), 1)}"
}

output "vault_unseal_key_name" {
  value = "${data.template_file.vault_unseal_key_name.rendered}"
}

output "instance_fqdns" {
  value = ["${aws_route53_record.per-instance.*.fqdn}"]
}

output "vault_security_group_id" {
  value = "${aws_security_group.vault.id}"
}

output "vault_instance_role_master" {  
  value = "${tarmak_vault_instance_role.master.init_token}"
}

output "vault_instance_role_worker" {  
  value = "${tarmak_vault_instance_role.worker.init_token}"
}

output "vault_instance_role_etcd" {  
  value = "${tarmak_vault_instance_role.etcd.init_token}"
}