data "aws_elb_hosted_zone_id" "main" {}

data "aws_caller_identity" "current" {
  provider = "aws"
}
