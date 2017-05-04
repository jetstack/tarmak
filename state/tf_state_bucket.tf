resource "aws_s3_bucket" "terraform_state" {
  count  = "${signum(length(var.bucket_prefix))}"
  bucket = "${var.bucket_prefix}${var.environment}-${var.region}-terraform-state"
  acl    = "private"

  tags {
    Description = "Terraform states for environment ${var.environment}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }

  versioning {
    enabled = true
  }
}

resource "aws_dynamodb_table" "terraform_state" {
  count          = "${signum(length(var.bucket_prefix))}"
  name           = "${var.bucket_prefix}${var.environment}-${var.region}-terraform-state"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  tags {
    Description = "Terraform state locks for environment ${var.environment}"
  }

  tags {
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}
