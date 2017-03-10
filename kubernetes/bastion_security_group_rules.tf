resource "aws_security_group_rule" "kubernetes_master_allow_bastion_ssh" {
  type                     = "ingress"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_master.id}"
  security_group_id        = "${data.terraform_remote_state.hub_tools.bastion_security_group_id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_bastion_ssh" {
  type                     = "ingress"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${data.terraform_remote_state.hub_tools.bastion_security_group_id}"
}

resource "aws_security_group_rule" "etcd_allow_bastion_ssh" {
  type                     = "ingress"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.etcd.id}"
  security_group_id        = "${data.terraform_remote_state.hub_tools.bastion_security_group_id}"
}

