data "aws_s3_bucket_object" "vault_ca" {
  bucket = "${data.terraform_remote_state.hub_network.secrets_bucket}"
  key    = "vault-${var.environment}/ca.pem"
}
