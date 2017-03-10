resource "aws_security_group" "etcd" {
  name   = "${data.template_file.stack_name.rendered}-k8s-etcd"
  vpc_id = "${data.terraform_remote_state.network.vpc_id}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-k8s-etcd"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_security_group_rule" "etcd_allow_ssh_CI" {
  type                     = "ingress"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
  source_security_group_id = "${data.terraform_remote_state.hub_tools.jenkins_security_group_id}"
  security_group_id        = "${aws_security_group.etcd.id}"
}

resource "aws_security_group_rule" "etcd_allow_etcd_client_kubernetes_master" {
  type                     = "ingress"
  from_port                = 2379
  to_port                  = 2379
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_master.id}"
  security_group_id        = "${aws_security_group.etcd.id}"
}

resource "aws_security_group_rule" "etcd_allow_etcd_client_events_kubernetes_master" {
  type                     = "ingress"
  from_port                = 2369
  to_port                  = 2369
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_master.id}"
  security_group_id        = "${aws_security_group.etcd.id}"
}

resource "aws_security_group_rule" "etcd_allow_etcd_client_cni_kubernetes_master" {
  type                     = "ingress"
  from_port                = 2359
  to_port                  = 2359
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_master.id}"
  security_group_id        = "${aws_security_group.etcd.id}"
}

resource "aws_security_group_rule" "etcd_allow_etcd_client_cni_kubernetes_worker" {
  type                     = "ingress"
  from_port                = 2359
  to_port                  = 2359
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.etcd.id}"
}

resource "aws_security_group_rule" "etcd_allow_etcd_peer_kubernetes_etcd" {
  type                     = "ingress"
  from_port                = 2379
  to_port                  = 2380
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.etcd.id}"
  security_group_id        = "${aws_security_group.etcd.id}"
}

resource "aws_security_group_rule" "etcd_allow_etcd_peer_events_etcd" {
  type                     = "ingress"
  from_port                = 2369
  to_port                  = 2370
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.etcd.id}"
  security_group_id        = "${aws_security_group.etcd.id}"
}

resource "aws_security_group_rule" "etcd_allow_etcd_peer_cni_etcd" {
  type                     = "ingress"
  from_port                = 2359
  to_port                  = 2360
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.etcd.id}"
  security_group_id        = "${aws_security_group.etcd.id}"
}

resource "aws_security_group_rule" "etcd_allow_node_exporter_check_private_net" {
  type                     = "ingress"
  from_port                = 9100
  to_port                  = 9100
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.etcd.id}"
}

resource "aws_security_group_rule" "etcd_allow_etcd_check_private_net" {
  type                     = "ingress"
  from_port                = 9115
  to_port                  = 9115
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.etcd.id}"
}

resource "aws_security_group_rule" "etcd_allow_egress" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.etcd.id}"
}
