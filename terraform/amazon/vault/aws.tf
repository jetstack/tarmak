provider "aws" {
  region              = "${var.region}"
  allowed_account_ids = ["${var.allowed_account_ids}"]
}

data "aws_caller_identity" "current" {
  provider = "aws"
}
