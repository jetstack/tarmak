provider "aws" {
  region = "${var.region}"
}

data "aws_elb_hosted_zone_id" "main" {}
