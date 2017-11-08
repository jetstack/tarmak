data "aws_subnet" "public" {
  count  = "${length(split(",", var.public_subnets))}"
  vpc_id = "${var.vpc_id}"
  id     = "${element(split(",", var.public_subnets), count.index)}"
}

data "aws_subnet" "private" {
  count  = "${length(split(",", var.private_subnets))}"
  vpc_id = "${var.vpc_id}"
  id     = "${element(split(",", var.private_subnets), count.index)}"
}
