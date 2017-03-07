resource "aws_s3_bucket" "terraform_state" {
  count  = "${signum(length(var.bucket_prefix))}"
  bucket = "${var.bucket_prefix}${var.environment}-${var.region}-terraform-state"
  acl    = "private"

  tags {
    Description = "Terraform states for environment ${var.environment}"
  }

  versioning {
    enabled = true
  }
}
