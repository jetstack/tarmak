resource "aws_security_group" "kubernetes_master" {
  name   = "${data.template_file.stack_name.rendered}-k8s-master"
  vpc_id = "${data.terraform_remote_state.network.vpc_id}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-k8s-master"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_security_group" "kubernetes_master_elb" {
  name   = "${data.template_file.stack_name.rendered}-k8s-master-elb"
  vpc_id = "${data.terraform_remote_state.network.vpc_id}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-k8s-master-elb"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_security_group_rule" "kubernetes_master_allow_ssh_CI" {
  type                     = "ingress"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
  source_security_group_id = "${data.terraform_remote_state.hub_tools.jenkins_security_group_id}"
  security_group_id        = "${aws_security_group.kubernetes_master.id}"
}

resource "aws_security_group_rule" "kubernetes_master_allow_bgp_kubernetes_master" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  source_security_group_id = "${aws_security_group.kubernetes_master.id}"
  security_group_id        = "${aws_security_group.kubernetes_master.id}"
}

resource "aws_security_group_rule" "kubernetes_master_allow_bgp_kubernetes_worker" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.kubernetes_master.id}"
}

resource "aws_security_group_rule" "kubernetes_master_elb_allow_6443_worker" {
  type                     = "ingress"
  from_port                = 6443
  to_port                  = 6443
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.kubernetes_master_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_master_elb_allow_6443_master" {
  type                     = "ingress"
  from_port                = 6443
  to_port                  = 6443
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_master.id}"
  security_group_id        = "${aws_security_group.kubernetes_master_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_master_elb_allow_6443_bastion" {
  type                     = "ingress"
  from_port                = 6443
  to_port                  = 6443
  protocol                 = "tcp"
  source_security_group_id = "${data.terraform_remote_state.hub_tools.bastion_security_group_id}"
  security_group_id        = "${aws_security_group.kubernetes_master_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_master_elb_allow_6443_CI" {
  type                     = "ingress"
  from_port                = 6443
  to_port                  = 6443
  protocol                 = "tcp"
  source_security_group_id = "${data.terraform_remote_state.hub_tools.jenkins_security_group_id}"
  security_group_id        = "${aws_security_group.kubernetes_master_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_master_allow_6443_kubernetes_master_elb" {
  type                     = "ingress"
  from_port                = 6443
  to_port                  = 6443
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_master_elb.id}"
  security_group_id        = "${aws_security_group.kubernetes_master.id}"
}

resource "aws_security_group_rule" "kubernetes_master_allow_6443_egress_kubernetes_master_elb" {
  type                     = "egress"
  from_port                = 6443
  to_port                  = 6443
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_master.id}"
  security_group_id        = "${aws_security_group.kubernetes_master_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_master_allow_egress" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.kubernetes_master.id}"
}
