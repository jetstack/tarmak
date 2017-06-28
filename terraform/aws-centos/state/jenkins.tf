variable "jenkins_data_size" {
  default = 40
}

resource "aws_ebs_volume" "jenkins" {
  count             = "${signum(length(var.jenkins_data_size))}"
  availability_zone = "${var.availability_zones[0]}"
  size              = "${var.jenkins_data_size}"
  type              = "gp2"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-jenkins"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}
