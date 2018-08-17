resource "aws_s3_bucket_object" "puppet-tar-gz" {
  key          = "${data.template_file.stack_name.rendered}/puppet.tar.gz"
  bucket       = "${var.secrets_bucket}"
  content_type = "application/tar+gzip"
  source       = "puppet.tar.gz"
  etag         = "${md5(file("puppet.tar.gz"))}"
}
