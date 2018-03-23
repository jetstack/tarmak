output "bastion_instance_id" {
  value = "${aws_instance.bastion.id}"
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
  value = "${aws_security_group.bastion.id}"
}

output "remote_admin_security_group_id" {
  value = "${aws_security_group.remote_admin.id}"
}