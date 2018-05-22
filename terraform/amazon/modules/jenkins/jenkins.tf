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


data "aws_subnet" "private_subnet" {
  id = "${var.private_subnet_ids[0]}"
}