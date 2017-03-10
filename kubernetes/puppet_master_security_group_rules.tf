resource "aws_security_group_rule" "puppet_master_allow_puppet_kubernetes_worker" {
  type                     = "ingress"
  from_port                = 8140
  to_port                  = 8140
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${data.terraform_remote_state.hub_tools.puppet_master_security_group_id}"
}

resource "aws_security_group_rule" "puppet_master_allow_puppet_kubernetes_master" {
  type                     = "ingress"
  from_port                = 8140
  to_port                  = 8140
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_master.id}"
  security_group_id        = "${data.terraform_remote_state.hub_tools.puppet_master_security_group_id}"
}

resource "aws_security_group_rule" "puppet_master_allow_puppet_etcd" {
  type                     = "ingress"
  from_port                = 8140
  to_port                  = 8140
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.etcd.id}"
  security_group_id        = "${data.terraform_remote_state.hub_tools.puppet_master_security_group_id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ssh_puppet_master" {
  type                     = "ingress"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
  source_security_group_id = "${data.terraform_remote_state.hub_tools.puppet_master_security_group_id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_master_allow_ssh_puppet_master" {
  type                     = "ingress"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
  source_security_group_id = "${data.terraform_remote_state.hub_tools.puppet_master_security_group_id}"
  security_group_id        = "${aws_security_group.kubernetes_master.id}"
}

resource "aws_security_group_rule" "etcd_allow_ssh_puppet_master" {
  type                     = "ingress"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
  source_security_group_id = "${data.terraform_remote_state.hub_tools.puppet_master_security_group_id}"
  security_group_id        = "${aws_security_group.etcd.id}"
}
