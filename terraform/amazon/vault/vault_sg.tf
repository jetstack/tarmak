resource "aws_security_group" "vault" {
  name   = "${data.template_file.stack_name.rendered}-vault"
  vpc_id = "${data.terraform_remote_state.network.vpc_id}"

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
  security_group_id = "${aws_security_group.vault.id}"
}

resource "aws_security_group_rule" "vault_in_allow_ssh_bastion" {
  type      = "ingress"
  from_port = 22
  to_port   = 22
  protocol  = "tcp"

  source_security_group_id = "${data.terraform_remote_state.tools.bastion_security_group_id}"
  security_group_id        = "${aws_security_group.vault.id}"
}

resource "aws_security_group_rule" "vault_in_allow_vault_bastion" {
  type      = "ingress"
  from_port = 8200
  to_port   = 8200
  protocol  = "tcp"

  source_security_group_id = "${data.terraform_remote_state.tools.bastion_security_group_id}"
  security_group_id        = "${aws_security_group.vault.id}"
}

resource "aws_security_group_rule" "vault_in_allow_everything_inner_cluster" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  security_group_id        = "${aws_security_group.vault.id}"
  source_security_group_id = "${aws_security_group.vault.id}"
}

output "vault_security_group_id" {
  value = "${aws_security_group.vault.id}"
}
