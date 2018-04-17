output "vpc_id" {
  value = "${data.aws_vpc.main.id}"
}

output "vpc_net" {
  value = "${data.aws_vpc.main.cidr_block}"
}

output "stack_name" {
  value = "${data.template_file.stack_name.rendered}"
}

output "private_subnet_ids" {
  value = ["${data.aws_subnet.private.*.id}"]
}

output "private_subnets" {
  value = ["${data.aws_subnet.private.*.cidr_block}"]
}

output "public_subnet_ids" {
  value = ["${data.aws_subnet.public.*.id}"]
}

output "public_subnets" {
  value = ["${data.aws_subnet.public.*.cidr_block}"]
}

output "route_table_public_ids" {
  value = ["${data.aws_route_table.public.*.id}"]
}

output "route_table_private_ids" {
  value = ["${data.aws_route_table.private.*.id}"]
}

output "private_zone_id" {
  value = ["${aws_route53_zone.private.*.id}"]
}

output "private_zone" {
  value = ["${aws_route53_zone.private.*.name}"]
}

output "environment" {
  value = "${var.environment}"
}

output "availability_zones" {
  value = "${var.availability_zones}"
}
