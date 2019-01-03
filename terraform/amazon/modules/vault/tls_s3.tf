resource "aws_s3_bucket_object" "node-certs" {
  count        = "${var.vault_min_instance_count}"
  key          = "vault/vault-${count.index+1}.pem-${md5(element(tls_locally_signed_cert.vault.*.cert_pem, count.index))}"
  bucket       = "${var.secrets_bucket}"
  content      = "${element(tls_locally_signed_cert.vault.*.cert_pem, count.index)}"
  content_type = "text/plain"
  kms_key_id   = "${var.vault_kms_key_id}"
}

resource "aws_s3_bucket_object" "node-keys" {
  count        = "${var.vault_min_instance_count}"
  key          = "vault/vault-${count.index+1}-key.pem-${md5(element(tls_private_key.vault.*.private_key_pem, count.index))}"
  bucket       = "${var.secrets_bucket}"
  content      = "${element(tls_private_key.vault.*.private_key_pem, count.index)}"
  content_type = "text/plain"
  kms_key_id   = "${var.vault_kms_key_id}"
}

resource "aws_s3_bucket_object" "ca-cert" {
  key          = "vault/ca.pem-${md5(tls_self_signed_cert.ca.0.cert_pem)}"
  bucket       = "${var.secrets_bucket}"
  kms_key_id   = "${var.vault_kms_key_id}"
  content      = "${tls_self_signed_cert.ca.0.cert_pem}"
  content_type = "text/plain"
}
