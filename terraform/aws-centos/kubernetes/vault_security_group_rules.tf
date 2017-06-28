resource "aws_security_group_rule" "vault_allow_vault_kubernetes_worker" {
  type                     = "ingress"
  from_port                = 8200
  to_port                  = 8200
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${data.terraform_remote_state.hub_vault.vault_security_group_id}"
}

resource "aws_security_group_rule" "vault_allow_vault_kubernetes_master" {
  type                     = "ingress"
  from_port                = 8200
  to_port                  = 8200
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_master.id}"
  security_group_id        = "${data.terraform_remote_state.hub_vault.vault_security_group_id}"
}

resource "aws_security_group_rule" "vault_allow_vault_etcd" {
  type                     = "ingress"
  from_port                = 8200
  to_port                  = 8200
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.etcd.id}"
  security_group_id        = "${data.terraform_remote_state.hub_vault.vault_security_group_id}"
}
