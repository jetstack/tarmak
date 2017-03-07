variable "puppet_master_instance_type" {
  default = "c4.large"
}

variable "puppet_master_root_size" {
  default = 20
}

variable "puppet_master_data_size" {
  default = 40
}

variable "puppet_deploy_key" {}

variable "foreman_admin_user" {
  default = "admin"
}

variable "foreman_admin_password" {}

resource "aws_security_group" "puppet_master" {
  name        = "${data.template_file.stack_name.rendered}-puppet_master"
  vpc_id      = "${data.terraform_remote_state.network.vpc_id}"
  description = "puppet master instance in ${data.template_file.stack_name.rendered}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-puppet_master"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_security_group_rule" "puppet_master_egress_allow_all" {
  type              = "egress"
  protocol          = -1
  from_port         = 0
  to_port           = 65535
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.puppet_master.id}"
}

resource "aws_security_group_rule" "puppet_master_ingress_allow_bastion_all" {
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = "${aws_security_group.bastion.id}"
  security_group_id        = "${aws_security_group.puppet_master.id}"
}

resource "aws_security_group_rule" "puppet_master_ingress_allow_jenkins_ssh" {
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 22
  to_port                  = 22
  source_security_group_id = "${aws_security_group.jenkins.id}"
  security_group_id        = "${aws_security_group.puppet_master.id}"
}

data "template_file" "puppet_master_user_data" {
  template = "${file("${path.module}/templates/puppet_master_user_data.yaml")}"

  vars {
    region                 = "${var.region}"
    dns_zone               = "${data.terraform_remote_state.network.private_zones[0]}"
    puppet_deploy_key      = "${var.puppet_deploy_key}"
    foreman_admin_user     = "${var.foreman_admin_user}"
    foreman_admin_password = "${var.foreman_admin_password}"
  }
}

resource "aws_instance" "puppet_master" {
  ami                    = "${var.centos_ami[var.region]}"
  instance_type          = "${var.puppet_master_instance_type}"
  subnet_id              = "${data.terraform_remote_state.network.private_subnet_ids[0]}"
  key_name               = "${var.key_name}"
  vpc_security_group_ids = ["${aws_security_group.puppet_master.id}"]
  iam_instance_profile   = "${aws_iam_role.puppet_master.name}"

  root_block_device = {
    volume_type = "gp2"
    volume_size = "${var.puppet_master_root_size}"
  }

  tags {
    Name        = "${data.template_file.stack_name.rendered}-puppet_master"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }

  user_data = "${data.template_file.puppet_master_user_data.rendered}"
}

resource "aws_ebs_volume" "puppet_master" {
  availability_zone = "${data.terraform_remote_state.network.availability_zones[0]}"
  size              = "${var.puppet_master_data_size}"
  type              = "gp2"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-puppet_master"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }

  lifecycle = {
    prevent_destroy = true
  }
}

resource "aws_volume_attachment" "puppet_master" {
  device_name  = "/dev/xvdd"
  volume_id    = "${aws_ebs_volume.puppet_master.id}"
  instance_id  = "${aws_instance.puppet_master.id}"
  skip_destroy = true
}

resource "aws_route53_record" "puppet_master" {
  zone_id = "${data.terraform_remote_state.network.private_zone_ids[0]}"
  name    = "puppet"
  type    = "A"
  ttl     = "300"
  records = ["${aws_instance.puppet_master.private_ip}"]
}

output "puppet_master_fqdn" {
  value = "${aws_route53_record.puppet_master.fqdn}"
}

output "puppet_master_security_group_id" {
  value = "${aws_security_group.puppet_master.id}"
}
