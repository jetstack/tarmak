resource "aws_s3_bucket_object" "puppet-tar-gz" {
  key          = "${data.template_file.stack_name.rendered}/puppet-manifests/${md5(file("puppet.tar.gz"))}-puppet.tar.gz"
  bucket       = "${var.secrets_bucket}"
  content_type = "application/tar+gzip"
  source       = "puppet.tar.gz"
  kms_key_id   = "${var.vault_kms_key_id}"
}

resource "aws_s3_bucket_object" "latest-puppet-hash" {
  key          = "${data.template_file.stack_name.rendered}/puppet-manifests/latest-puppet-hash"
  bucket       = "${var.secrets_bucket}"
  content_type = "application/tar+gzip"
  content      = "${md5(file("puppet.tar.gz"))}"
}

resource "aws_s3_bucket_object" "legacy-puppet-tar-gz" {
  key          = "${data.template_file.stack_name.rendered}/puppet.tar.gz"
  bucket       = "${var.secrets_bucket}"
  content_type = "application/tar+gzip"
  source       = "puppet.tar.gz"
  kms_key_id   = "${var.vault_kms_key_id}"
}
