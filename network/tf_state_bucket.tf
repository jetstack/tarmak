resource "aws_s3_bucket" "terraform_state" {
  count  = "${length(var.state_buckets)}"
  bucket = "${element(var.state_buckets, count.index)}-${var.environment}-${var.region}-terraform-state"
  acl    = "private"

  tags {
    Description = "Bucket to store terraform states"
  }

  versioning {
    enabled = true
  }
}
