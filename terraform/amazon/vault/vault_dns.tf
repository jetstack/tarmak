resource "aws_route53_record" "per-instance" {
  count   = "${var.instance_count}"
  zone_id = "${data.terraform_remote_state.network.private_zone_id.0}"
  name    = "vault-${count.index + 1}"
  type    = "A"
  ttl     = "180"
  records = ["${element(aws_instance.vault.*.private_ip, count.index)}"]
}

resource "aws_route53_record" "endpoint" {
  zone_id = "${data.terraform_remote_state.network.private_zone_id.0}"
  name    = "vault"
  type    = "A"
  ttl     = "180"
  records = ["${aws_instance.vault.*.private_ip}"]
}

output "instance_fqdns" {
  value = ["${aws_route53_record.per-instance.*.fqdn}"]
}
