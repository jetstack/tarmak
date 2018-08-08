output "bastion_status" {
  value = "${data.tarmak_bastion_instance.bastion.status}"
}

output "bastion_security_group_id" {
  value = "${element(concat(aws_security_group.bastion.*.id, list("")), 0)}"
}

output "bastion_instance_id" {
  value = "${element(concat(aws_instance.bastion.*.id, list("")), 0)}"
}