resource "aws_security_group" "remote_admin" {
  name        = "${data.template_file.stack_name.rendered}-remote_admin"
  vpc_id      = "${data.terraform_remote_state.network.vpc_id}"
  description = "Allow remote admin access to nodes for ${data.template_file.stack_name.rendered}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-remote_admin"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_security_group_rule" "remote_admin_egress_allow_all" {
  type              = "egress"
  protocol          = -1
  from_port         = 0
  to_port           = 65535
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.remote_admin.id}"
}

resource "aws_security_group_rule" "remote_admin_ingress_allow_ssh_all" {
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 22
  to_port                  = 22
  security_group_id        = "${aws_security_group.remote_admin.id}"
  source_security_group_id = "${aws_security_group.bastion.id}"
}

output "remote_admin_security_group_id" {
  value = "${aws_security_group.remote_admin.id}"
}
