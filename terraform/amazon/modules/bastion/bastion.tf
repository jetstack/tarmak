data "template_file" "bastion_user_data" {
  template = "${file("${path.module}/templates/bastion_user_data.yaml")}"

  vars {
    fqdn = "bastion.${var.public_zone}"
  }
}

data "template_file" "stack_name" {
  template = "${var.stack_name_prefix}${var.environment}-${var.name}"
}

resource "aws_security_group" "bastion" {
  count       = 1
  name        = "${data.template_file.stack_name.rendered}-bastion"
  vpc_id      = "${var.vpc_id}"
  description = "Bastion instance in ${data.template_file.stack_name.rendered}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-bastion"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

data "tarmak_bastion_instance" "bastion" {
  hostname    = "bastion"
  username    = "centos"

  depends_on = ["aws_instance.bastion"]
}

resource "aws_instance" "bastion" {
  ami                    = "${var.bastion_ami}"
  instance_type          = "${var.bastion_instance_type}"
  subnet_id              = "${var.public_subnet_ids[0]}"
  key_name               = "${var.key_name}"
  vpc_security_group_ids = ["${aws_security_group.bastion.0.id}"]

  root_block_device = {
    volume_type = "gp2"
    volume_size = "${var.bastion_root_size}"
  }

  tags {
    Name        = "${data.template_file.stack_name.rendered}-bastion"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
    tarmak_role = "bastion"
  }

  user_data = "${data.template_file.bastion_user_data.rendered}"
}

resource "aws_eip" "bastion" {
  vpc      = true
  instance = "${aws_instance.bastion.0.id}"
}

resource "aws_security_group_rule" "egress_allow_all" {
  type              = "egress"
  protocol          = -1
  from_port         = 0
  to_port           = 65535
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.bastion.0.id}"
}

resource "aws_security_group_rule" "ingress_allow_ssh_all" {
  type              = "ingress"
  protocol          = "tcp"
  from_port         = 22
  to_port           = 22
  cidr_blocks       = ["${var.bastion_admin_cidrs}"]
  security_group_id = "${aws_security_group.bastion.0.id}"
}

resource "aws_route53_record" "bastion" {
  zone_id = "${var.public_zone_id}"
  name    = "bastion.${var.environment}"
  type    = "A"
  ttl     = "300"
  records = ["${aws_eip.bastion.public_ip}"]
}

resource "aws_route53_record" "bastion_private" {
  zone_id = "${var.private_zone_id}"
  name    = "bastion.${var.environment}"
  type    = "A"
  ttl     = "60"
  records = ["${aws_instance.bastion.0.private_ip}"]
}
