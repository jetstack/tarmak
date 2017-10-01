provider "aws" {
  region              = "${var.region}"
  allowed_account_ids = ["${var.allowed_account_ids}"]
  version             = "~> 1.0"
}

provider "template" {
  version = "~> 1.0"
}

provider "terraform" {
  version = "~> 1.0"
}

data "aws_caller_identity" "current" {}

data "aws_elb_hosted_zone_id" "main" {}
