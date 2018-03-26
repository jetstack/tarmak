resource "aws_route53_record" "per-instance" {
  count   = "${var.vault_instance_count}"
  zone_id = "${var.private_zone_id}"
  name    = "vault-${count.index + 1}"
  type    = "A"
  ttl     = "180"
  records = ["${element(aws_instance.vault.*.private_ip, count.index)}"]
}

resource "aws_route53_record" "endpoint" {
  zone_id = "${var.private_zone_id}"
  name    = "vault"
  type    = "A"
  ttl     = "180"
  records = ["${aws_instance.vault.*.private_ip}"]
}
