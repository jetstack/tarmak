resource "aws_s3_bucket" "vault-backup" {
  bucket = "${data.terraform_remote_state.network.bucket_prefix}${var.environment}-${var.region}-vault-backup"
  acl    = "private"

  lifecycle_rule {
    id      = "backup"
    prefix  = ""
    enabled = true

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    expiration {
      days = 365
    }
  }

  tags {
    Description = "Vault backups for ${var.environment}"
  }
}
