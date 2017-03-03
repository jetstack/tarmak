resource "aws_route53_zone" "private" {
  count   = "${length(var.private_zones)}"
  name    = "${var.private_zones[count.index]}"
  vpc_id  = "${aws_vpc.main.id}"
  comment = "Hosted zone for private kubernetes in ${var.environment}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_route53_zone" "public" {
  count   = "${length(var.public_zones)}"
  name    = "${var.public_zones[count.index]}"
  comment = "Hosted zone for public kubernetes in ${var.environment}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}
