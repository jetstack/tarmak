resource "aws_route53_record" "star-txt" {
  zone_id = "${var.public_zone_id}"
  name    = "*._tarmak.${var.environment}"
  type    = "TXT"
  ttl     = "300"
  records = ["tarmak delegation works"]
}
