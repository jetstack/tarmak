resource "aws_security_group" "vault" {
  count  = 1
  name   = "${data.template_file.stack_name.rendered}-vault"
  vpc_id = "${var.vpc_id}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-vault"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_security_group_rule" "vault_out_allow_all" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.vault.0.id}"
}

resource "aws_security_group_rule" "vault_in_allow_ssh_bastion" {
  type      = "ingress"
  from_port = 22
  to_port   = 22
  protocol  = "tcp"

  source_security_group_id = "${var.bastion_security_group_id}"
  security_group_id        = "${aws_security_group.vault.0.id}"
}

resource "aws_security_group_rule" "vault_in_allow_vault_bastion" {
  type      = "ingress"
  from_port = 8200
  to_port   = 8200
  protocol  = "tcp"

  source_security_group_id = "${var.bastion_security_group_id}"
  security_group_id        = "${aws_security_group.vault.0.id}"
}

resource "aws_security_group_rule" "vault_in_allow_spire_bastion" {
  type      = "ingress"
  from_port = 8081
  to_port   = 8081
  protocol  = "tcp"

  source_security_group_id = "${var.bastion_security_group_id}"
  security_group_id        = "${aws_security_group.vault.0.id}"
}

resource "aws_security_group_rule" "vault_in_allow_everything_inner_cluster" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  security_group_id        = "${aws_security_group.vault.0.id}"
  source_security_group_id = "${aws_security_group.vault.0.id}"
}
