data "terraform_remote_state" "vpc_peer_stack" {
  count   = "${signum(length(var.vpc_peer_stack))}"
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "network_${var.environment}_${var.vpc_peer_stack}.tfstate"
  }
}

resource "aws_route53_zone_association" "hub_zone" {
  count   = "${signum(length(var.vpc_peer_stack))}"
  zone_id = "${data.terraform_remote_state.vpc_peer_stack.private_zone_id}"
  vpc_id  = "${var.vpc_id}"
}
