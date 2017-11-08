resource "aws_internet_gateway" "main" {
  vpc_id = "${var.vpc_id}"

  tags {
    Name        = "vpc.${data.template_file.stack_name.rendered}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}
