provider "aws" {
  region              = "${var.region}"
  allowed_account_ids = ["${var.allowed_account_ids}"]
}

data "aws_acm_certificate" "wildcard" {
  domain   = "*.${data.terraform_remote_state.network.public_zones[0]}"
  statuses = ["ISSUED"]
}

data "aws_elb_hosted_zone_id" "main" {}
