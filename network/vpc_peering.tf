data "terraform_remote_state" "vpc_peer_stack" {
  count   = "${signum(length(var.vpc_peer_stack))}"
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "network_${var.environment}_${var.vpc_peer_stack}.tfstate"
  }
}

resource "aws_vpc_peering_connection" "peerings" {
  count       = "${signum(length(var.vpc_peer_stack))}"
  count       = "${length(var.vpc_peers)}"
  peer_vpc_id = "${data.terraform_remote_state.vpc_peer_stack.vpc_id}"
  vpc_id      = "${aws_vpc.main.id}"
  auto_accept = true
}
