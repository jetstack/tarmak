resource "aws_s3_bucket_object" "puppet-tar-gz" {
  key          = "${data.template_file.stack_name.rendered}/puppet.tar.gz"
  bucket       = "${var.secrets_bucket}"
  content_type = "application/tar+gzip"
  source       = "puppet.tar.gz"
  etag         = "82cb760fb0d10f4813d675adcd09233a"
  # TODO: etag         = "${md5(file("../../puppet.tar.gz"))}"
}
