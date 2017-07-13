resource "aws_route53_zone" "public" {
  count   = "${signum(length(var.public_zone))}"
  name    = "${var.public_zone}"
  comment = "Hosted zone for public kubernetes in ${var.environment}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_route53_record" "star-txt" {
  zone_id = "${aws_route53_zone.public.zone_id}"
  name    = "*._tarmak"
  type    = "TXT"
  ttl     = "300"
  records = ["tarmak delegation works"]
}
