output "vpc_id" {
  value = "${aws_vpc.main.id}"
}

output "vpc_name" {
  value = "${var.vpc_name}"
}

output "vpc_cidr_block" {
  value = "${aws_vpc.main.cidr_block}"
}

output "private_subnet_ids" {
  value = ["${aws_subnet.private.*.id}"]
}

output "public_subnet_ids" {
  value = ["${aws_subnet.public.*.id}"]
}

output "nat_public_ips" {
  value = ["${aws_nat_gateway.main.*.public_ip}"]
}

output "route_table_public_ids" {
  value = ["${aws_route_table.public.*.id}"]
}

output "route_table_private_ids" {
  value = ["${aws_route_table.private.*.id}"]
}

output "environment" {
  value = "${var.environment}"
}

output "availability_zones" {
  value = "${var.availability_zones}"
}

output "aws_route53_zone_private_id" {
  value = ["${aws_route53_zone.private.*.id}"]
}

output "aws_route53_zone_public_id" {
  value = ["${aws_route53_zone.public.*.id}"]
}
