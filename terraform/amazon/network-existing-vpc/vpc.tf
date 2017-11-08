resource "aws_vpc" "main" {
  cidr_block           = "${var.network}"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags {
    Name        = "vpc.${data.template_file.stack_name.rendered}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}
