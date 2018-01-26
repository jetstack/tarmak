resource "aws_s3_bucket_object" "node-certs" {
  count        = "${var.instance_count}"
  key          = "vault/vault-${count.index+1}.pem-${md5(element(tls_locally_signed_cert.vault.*.cert_pem, count.index))}"
  bucket       = "${data.terraform_remote_state.state.secrets_bucket.0}"
  content      = "${element(tls_locally_signed_cert.vault.*.cert_pem, count.index)}"
  content_type = "text/plain"
  kms_key_id   = "${data.terraform_remote_state.state.secrets_kms_arn.0}"
}

resource "aws_s3_bucket_object" "node-keys" {
  count        = "${var.instance_count}"
  key          = "vault/vault-${count.index+1}-key.pem-${md5(element(tls_private_key.vault.*.private_key_pem, count.index))}"
  bucket       = "${data.terraform_remote_state.state.secrets_bucket.0}"
  content      = "${element(tls_private_key.vault.*.private_key_pem, count.index)}"
  content_type = "text/plain"
  kms_key_id   = "${data.terraform_remote_state.state.secrets_kms_arn.0}"
}

resource "aws_s3_bucket_object" "ca-cert" {
  key          = "vault/ca.pem-${md5(tls_self_signed_cert.ca.cert_pem)}"
  bucket       = "${data.terraform_remote_state.state.secrets_bucket.0}"
  kms_key_id   = "${data.terraform_remote_state.state.secrets_kms_arn.0}"
  content      = "${tls_self_signed_cert.ca.cert_pem}"
  content_type = "text/plain"
}
