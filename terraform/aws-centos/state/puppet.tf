variable "puppet_master_data_size" {
  default = 40
}

resource "aws_ebs_volume" "puppet_master" {
  count             = "${signum(length(var.puppet_master_data_size))}"
  availability_zone = "${var.availability_zones[0]}"
  size              = "${var.puppet_master_data_size}"
  type              = "gp2"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-puppet_master"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}
