output "bastion_instance_id" {
  value = "${element(concat(aws_instance.bastion.*.id, list("")), 0)}"
}


output "bastion_fqdn" {
  value = "${aws_route53_record.bastion.fqdn}"
}

output "bastion_private_ip" {
  value = "${aws_eip.bastion.public_ip}"
}

output "bastion_ip" {
  value = "${aws_eip.bastion.public_ip}"
}

output "bastion_security_group_id" {
  value = "${element(concat(aws_security_group.bastion.*.id, list("")), 0)}"
}

output "remote_admin_security_group_id" {
  value = "${aws_security_group.remote_admin.id}"
}