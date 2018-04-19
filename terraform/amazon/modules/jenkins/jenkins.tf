data "template_file" "stack_name" {
  template = "${var.stack_name_prefix}${var.environment}-${var.name}"
}

resource "aws_security_group" "jenkins" {
  name        = "${data.template_file.stack_name.rendered}-jenkins"
  vpc_id      = "${var.vpc_id}"
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
  source_security_group_id = "${var.bastion_security_group_id}"
  security_group_id        = "${aws_security_group.jenkins.id}"
}

data "template_file" "jenkins_user_data" {
  template = "${file("${path.module}/templates/jenkins_user_data.yaml")}"

  vars {
    region = "${var.region}"
    fqdn   = "jenkins.${var.private_zone}"
  }
}

resource "aws_instance" "jenkins" {
  ami                    = "${var.jenkins_ami}"
  instance_type          = "${var.jenkins_instance_type}"
  subnet_id              = "${var.private_subnet_ids[0]}"
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

  lifecycle {
    ignore_changes = ["volume_tags"]
  }
}

resource "aws_volume_attachment" "jenkins" {
  device_name = "/dev/xvdd"
  volume_id   = "${aws_ebs_volume.jenkins.id}"
  instance_id = "${aws_instance.jenkins.id}"
  skip_destroy = true
}

resource "aws_ebs_volume" "jenkins" {
    availability_zone = "${data.aws_subnet.private_subnet.availability_zone}"
    encrypted = true
    size = "${var.jenkins_ebs_size}"
    tags {
        Name = "Jenkins"
    }
}

data "aws_subnet" "private_subnet" {
  id = "${var.private_subnet_ids[0]}"
}

resource "aws_route53_record" "jenkins_private" {
  zone_id = "${var.private_zone_id}"
  name    = "jenkins.${var.environment}"
  type    = "A"
  ttl     = "60"
  records = ["${aws_instance.jenkins.private_ip}"]
}