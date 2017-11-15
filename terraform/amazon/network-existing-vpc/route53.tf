resource "aws_route53_zone" "private" {
  count         = "${signum(length(var.private_zone))}"
  name          = "${var.private_zone}"
  vpc_id        = "${var.vpc_id}"
  force_destroy = true
  comment       = "Hosted zone for private kubernetes in ${var.environment}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}
