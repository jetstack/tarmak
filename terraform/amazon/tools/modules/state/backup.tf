variable "backup_expiration_days" {
  default = 365
}

variable "backup_transition_glacier_days" {
  default = 90
}

resource "aws_s3_bucket" "backups" {
  count  = "${signum(length(var.bucket_prefix))}"
  bucket = "${var.bucket_prefix}${var.environment}-${var.region}-backups"
  acl    = "private"

  tags {
    Description = "Backups for environment ${var.environment}"
  }

  force_destroy = "true"

  lifecycle_rule {
    prefix  = ""
    enabled = true

    transition {
      days          = "${var.backup_transition_glacier_days}"
      storage_class = "GLACIER"
    }

    expiration {
      days = "${var.backup_expiration_days}"
    }
  }
}

resource "aws_kms_key" "backups" {
  count                   = "${signum(length(var.bucket_prefix))}"
  description             = "Encrypts backups for environment ${var.environment}"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "backups" {
  name          = "alias/tarmak/${var.environment}/backups"
  target_key_id = "${aws_kms_key.backups.key_id}"
}

output "backups_bucket" {
  value = "${aws_s3_bucket.backups.*.bucket}"
}

output "backups_kms_arn" {
  value = "${aws_kms_key.backups.*.arn}"
}
