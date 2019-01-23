output "vpc_id" {
  value = "${element(concat(aws_vpc.main.*.id, list("")), 0)}"
}

output "vpc_net" {
  value = "${element(concat(aws_vpc.main.*.id, list("")), 0)}"
}

output "stack_name" {
  value = "${data.template_file.stack_name.rendered}"
}

output "private_subnet_ids" {
  value = ["${aws_subnet.private.*.id}"]
}

output "private_subnets" {
  value = ["${aws_subnet.private.*.cidr_block}"]
}

output "public_subnet_ids" {
  value = ["${aws_subnet.public.*.id}"]
}

output "public_subnets" {
  value = ["${aws_subnet.public.*.cidr_block}"]
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

output "private_zone_id" {
  value = "${concat(aws_route53_zone.private.*.id, list(""))}"
}

# remove trailing dots from the name
output "private_zone" {
  value = "${list(replace(aws_route53_zone.private.0.name, "/\\.$/", ""), "")}"
}

output "environment" {
  value = "${var.environment}"
}

output "availability_zones" {
  value = "${var.availability_zones}"
}
