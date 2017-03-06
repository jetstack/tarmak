variable "jenkins_instance_type" {
  default = "c4.large"
}

variable "jenkins_root_size" {
  default = 20
}

variable "jenkins_data_size" {
  default = 40
}

resource "aws_security_group" "jenkins" {
  name        = "${data.template_file.stack_name.rendered}-jenkins"
  vpc_id      = "${data.terraform_remote_state.network.vpc_id}"
  description = "Jenkins instance in ${data.template_file.stack_name.rendered}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-jenkins"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_security_group_rule" "jenkins_egress_allow_all" {
  type              = "egress"
  protocol          = -1
  from_port         = 0
  to_port           = 65535
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.jenkins.id}"
}

resource "aws_security_group_rule" "jenkins_ingress_allow_bastion_all" {
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = "${aws_security_group.bastion.id}"
  security_group_id        = "${aws_security_group.jenkins.id}"
}

data "template_file" "jenkins_user_data" {
  template = "${file("${path.module}/templates/jenkins_user_data.yaml")}"

  vars {
    region   = "${var.region}"
    dns_zone = "${data.terraform_remote_state.network.private_zones[0]}"
  }
}

resource "aws_instance" "jenkins" {
  ami                    = "${var.centos_ami[var.region]}"
  instance_type          = "${var.jenkins_instance_type}"
  subnet_id              = "${data.terraform_remote_state.network.private_subnet_ids[0]}"
  key_name               = "${var.key_name}"
  vpc_security_group_ids = ["${aws_security_group.jenkins.id}"]
  iam_instance_profile   = "${aws_iam_role.jenkins.name}"

  root_block_device = {
    volume_type = "gp2"
    volume_size = "${var.jenkins_root_size}"
  }

  tags {
    Name        = "${data.template_file.stack_name.rendered}-jenkins"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }

  user_data = "${data.template_file.jenkins_user_data.rendered}"
}

resource "aws_ebs_volume" "jenkins" {
  availability_zone = "${data.terraform_remote_state.network.availability_zones[0]}"
  size              = "${var.jenkins_data_size}"
  type              = "gp2"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-jenkins"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }

  lifecycle = {
    prevent_destroy = true
  }
}

resource "aws_volume_attachment" "jenkins" {
  device_name  = "/dev/xvdd"
  volume_id    = "${aws_ebs_volume.jenkins.id}"
  instance_id  = "${aws_instance.jenkins.id}"
  skip_destroy = true
}

resource "aws_route53_record" "jenkins" {
  zone_id = "${data.terraform_remote_state.network.private_zone_ids[0]}"
  name    = "jenkins"
  type    = "A"
  ttl     = "300"
  records = ["${aws_instance.jenkins.private_ip}"]
}

output "jenkins_fqdn" {
  value = "${aws_route53_record.jenkins.fqdn}"
}
