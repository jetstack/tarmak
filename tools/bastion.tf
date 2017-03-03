variable "bastion_instance_type" {
  default = "t2.micro"
}

resource "aws_eip" "bastion" {
  vpc      = true
  instance = "${aws_instance.bastion.id}"
}

resource "aws_security_group" "bastion" {
  name   = "${data.template_file.stack_name.rendered}-bastion"
  vpc_id = "${data.terraform_remote_state.network.vpc_id}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-bastion"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_security_group_rule" "egress_allow_all" {
  type              = "egress"
  protocol          = -1
  from_port         = 0
  to_port           = 65535
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.bastion.id}"
}

resource "aws_security_group_rule" "ingress_allow_ssh_admin" {
  type              = "ingress"
  protocol          = "tcp"
  from_port         = 22
  to_port           = 22
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.bastion.id}"
}

resource "aws_instance" "bastion" {
  ami                    = "${var.centos_ami[var.region]}"
  instance_type          = "${var.bastion_instance_type}"
  subnet_id              = "${data.terraform_remote_state.network.private_subnet_ids[0]}"
  key_name               = "${var.key_name}"
  vpc_security_group_ids = ["${aws_security_group.bastion.id}"]

  tags {
    Name        = "${data.template_file.stack_name.rendered}-bastion"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_route53_record" "bastion" {
  zone_id = "${data.terraform_remote_state.network.public_zone_ids[0]}"
  name    = "www"
  type    = "A"
  ttl     = "300"
  records = ["${aws_eip.bastion.public_ip}"]
}

output "bastion_fqdn" {
  value = "${aws_route53_record.bastion.fqdn}"
}

output "bastion_ip" {
  value = "${aws_eip.bastion.public_ip}"
}
