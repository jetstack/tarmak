output "stack_name" {
  value = "${data.template_file.stack_name.rendered}"
}

output "environment" {
  value = "${var.environment}"
}

output "public_zone_id" {
  value = "${var.public_zone_id}"
}

output "public_zone" {
  value = "${var.public_zone}"
}

output "bucket_prefix" {
  value = "${var.bucket_prefix}"
}

output "secrets_bucket" {
  value = "${aws_s3_bucket.secrets.*.bucket}"
}

output "secrets_kms_arn" {
  value = "${aws_kms_key.secrets.*.arn}"
}

