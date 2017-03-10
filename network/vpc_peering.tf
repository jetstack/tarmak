data "terraform_remote_state" "vpc_peer_stack" {
  count   = "${signum(length(var.vpc_peer_stack))}"
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "network_${var.environment}_${var.vpc_peer_stack}.tfstate"
  }
}

resource "aws_vpc_peering_connection" "peering" {
  count       = "${signum(length(var.vpc_peer_stack))}"
  peer_vpc_id = "${data.terraform_remote_state.vpc_peer_stack.vpc_id}"
  vpc_id      = "${aws_vpc.main.id}"
  auto_accept = true

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  requester {
    allow_remote_vpc_dns_resolution = true
  }
}

resource "aws_route" "myself_peering_private" {
  count                     = "${signum(length(var.vpc_peer_stack))*length(var.availability_zones)}"
  route_table_id            = "${aws_route_table.private.*.id[count.index]}"
  destination_cidr_block    = "${data.terraform_remote_state.vpc_peer_stack.vpc_net}"
  vpc_peering_connection_id = "${aws_vpc_peering_connection.peering.id}"
}

resource "aws_route" "myself_peering_public" {
  count                     = "${signum(length(var.vpc_peer_stack))}"
  route_table_id            = "${aws_route_table.public.id}"
  destination_cidr_block    = "${data.terraform_remote_state.vpc_peer_stack.vpc_net}"
  vpc_peering_connection_id = "${aws_vpc_peering_connection.peering.id}"
}

resource "aws_route" "them_peering_public" {
  count                     = "${signum(length(var.vpc_peer_stack))}"
  route_table_id            = "${data.terraform_remote_state.vpc_peer_stack.route_table_public_ids[0]}"
  destination_cidr_block    = "${var.network}"
  vpc_peering_connection_id = "${aws_vpc_peering_connection.peering.id}"
}

resource "aws_route" "them_peering_private" {
  count                     = "${signum(length(var.vpc_peer_stack))*length(var.availability_zones)}"
  route_table_id            = "${data.terraform_remote_state.vpc_peer_stack.route_table_private_ids[count.index]}"
  destination_cidr_block    = "${var.network}"
  vpc_peering_connection_id = "${aws_vpc_peering_connection.peering.id}"
}

resource "aws_route53_zone_association" "hub_zone" {
  count   = "${signum(length(var.vpc_peer_stack))}"
  zone_id = "${data.terraform_remote_state.vpc_peer_stack.private_zone_ids[0]}"
  vpc_id  = "${aws_vpc.main.id}"
}
