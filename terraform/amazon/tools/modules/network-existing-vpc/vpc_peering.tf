resource "aws_route53_zone_association" "hub_zone" {
  count   = "${signum(length(var.vpc_peer_stack))}"
  zone_id = "${var.private_zone_id}"
  vpc_id  = "${var.vpc_id}"
}
