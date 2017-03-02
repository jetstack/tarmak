resource "aws_subnet" "public" {
  count                   = "${length(var.availability_zones)}"
  vpc_id                  = "${aws_vpc.main.id}"
  cidr_block              = "${cidrsubnet(cidrsubnet(aws_vpc.main.cidr_block, 3, 0), 3, count.index)}"
  availability_zone       = "${var.availability_zones[count.index]}"
  map_public_ip_on_launch = true

  tags {
    Name        = "${var.vpc_name}_public_${var.availability_zones[count.index]}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_subnet" "private" {
  count             = "${length(var.availability_zones)}"
  vpc_id            = "${aws_vpc.main.id}"
  cidr_block        = "${cidrsubnet(aws_vpc.main.cidr_block, 3, count.index + 1)}"
  availability_zone = "${var.availability_zones[count.index]}"

  tags {
    Name        = "${var.vpc_name}_private_${var.availability_zones[count.index]}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}
