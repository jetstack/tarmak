resource "aws_vpc_peering_connection" "peering" {
  count       = "${signum(length(var.vpc_peer_stack))}"
  peer_vpc_id = "${var.peer_vpc_id}"
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
  destination_cidr_block    = "${var.vpc_net}"
  vpc_peering_connection_id = "${aws_vpc_peering_connection.peering.id}"
}

resource "aws_route" "myself_peering_public" {
  count                     = "${signum(length(var.vpc_peer_stack))}"
  route_table_id            = "${aws_route_table.public.id}"
  destination_cidr_block    = "${var.vpc_net}"
  vpc_peering_connection_id = "${aws_vpc_peering_connection.peering.id}"
}

resource "aws_route" "them_peering_public" {
  count                     = "${signum(length(var.vpc_peer_stack))}"
  route_table_id            = "${var.route_table_public_ids[0]}"
  destination_cidr_block    = "${var.network}"
  vpc_peering_connection_id = "${aws_vpc_peering_connection.peering.id}"
}

resource "aws_route" "them_peering_private" {
  count                     = "${signum(length(var.vpc_peer_stack))*length(var.availability_zones)}"
  route_table_id            = "${var.route_table_private_ids[count.index]}"
  destination_cidr_block    = "${var.network}"
  vpc_peering_connection_id = "${aws_vpc_peering_connection.peering.id}"
}

resource "aws_route53_zone_association" "hub_zone" {
  count   = "${signum(length(var.vpc_peer_stack))}"
  zone_id = "${var.private_zone_id}"
  vpc_id  = "${aws_vpc.main.id}"
}
