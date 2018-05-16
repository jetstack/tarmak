resource "aws_s3_bucket_object" "wing-binary" {
  source = "wing_linux_amd64"
  bucket = "${var.secrets_bucket}"
  key    = "${data.template_file.stack_name.rendered}/wing-${md5(file("wing_linux_amd64"))}"
  etag   = "${md5(file("wing_linux_amd64"))}"
}
