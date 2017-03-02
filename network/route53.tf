resource "aws_route53_zone" "private" {
  count  = 1
  name   = "int.todo.com"
  vpc_id = "${aws_vpc.main.id}"

  tags {
    Name        = "vpc.${var.vpc_name}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_route53_zone" "public" {
  count  = 1
  name   = "todo.com"
  vpc_id = "${aws_vpc.main.id}"

  tags {
    Name        = "vpc.${var.vpc_name}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}
