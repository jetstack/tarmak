resource "aws_s3_bucket" "secrets" {
  count  = "${length(var.secrets_buckets)}"
  bucket = "${element(var.secrets_buckets, count.index)}-${var.environment}-${var.region}-secrets"
  acl    = "private"

  tags {
    Description = "Secrets for environment ${var.environment}"
  }

  versioning {
    enabled = true
  }
}

resource "aws_kms_key" "secrets" {
  count                   = "${var.secrets_kms}"
  description             = "Encrypts secrets for environment ${var.environment}"
  deletion_window_in_days = 7
}

output "secrets_bucket" {
 value = "${aws_s3_bucket.secrets.bucket}"
}

output "secrets_kms_arn" {
  value = "${aws_kms_key.secrets.arn}"
}
