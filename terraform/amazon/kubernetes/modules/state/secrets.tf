resource "aws_s3_bucket" "secrets" {
  count  = "${signum(length(var.bucket_prefix))}"
  bucket = "${var.bucket_prefix}${var.environment}-${var.region}-secrets"
  acl    = "private"

  tags {
    Description = "Secrets for environment ${var.environment}"
  }

  versioning {
    enabled = true
  }
}

resource "aws_kms_key" "secrets" {
  count                   = "${signum(length(var.bucket_prefix))}"
  description             = "Encrypts secrets for environment ${var.environment}"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "secrets" {
  name          = "alias/tarmak/${var.environment}/secrets"
  target_key_id = "${aws_kms_key.secrets.key_id}"
}