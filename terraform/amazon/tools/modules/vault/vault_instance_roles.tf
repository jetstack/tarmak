resource "tarmak_vault_instance_role" "master" {  
  role_name = "master"
  vault_cluster_name = "${var.vault_cluster_name}"
  internal_fqdns = ["${aws_route53_record.per-instance.*.fqdn}"]
  vault_ca = "${tls_self_signed_cert.ca.cert_pem}"
}

resource "tarmak_vault_instance_role" "worker" {
  role_name = "worker"
  vault_cluster_name = "${var.vault_cluster_name}"
  internal_fqdns = ["${aws_route53_record.per-instance.*.fqdn}"]
  vault_ca = "${tls_self_signed_cert.ca.cert_pem}"
}

resource "tarmak_vault_instance_role" "etcd" {
  role_name = "etcd"
  vault_cluster_name = "${var.vault_cluster_name}"
  internal_fqdns = ["${aws_route53_record.per-instance.*.fqdn}"]
  vault_ca = "${tls_self_signed_cert.ca.cert_pem}"
}