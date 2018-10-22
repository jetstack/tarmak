/*provider "aws" {
  region              = "${var.region}"
  allowed_account_ids = ["${var.allowed_account_ids}"]
  version             = "~> 1.7.1"
}

provider "awstag" {
  region              = "${var.region}"
  allowed_account_ids = ["${var.allowed_account_ids}"]
}

provider "template" {
  version = "~> 1.0"
}
*/

data "aws_caller_identity" "current" {}

data "aws_elb_hosted_zone_id" "main" {}
