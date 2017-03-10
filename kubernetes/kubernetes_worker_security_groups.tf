resource "aws_security_group" "kubernetes_worker" {
  name   = "${data.template_file.stack_name.rendered}-k8s-worker"
  vpc_id = "${data.terraform_remote_state.network.vpc_id}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-k8s-worker"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"

    # Required for AWS cloud provider
    KubernetesCluster = "${data.template_file.stack_name.rendered}"
  }
}

resource "aws_security_group" "kubernetes_nodeport_elb" {
  name   = "${data.template_file.stack_name.rendered}-k8s-nodeport-elb"
  vpc_id = "${data.terraform_remote_state.network.vpc_id}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-k8s-nodeport-elb"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"

    # Required for AWS cloud provider
    KubernetesCluster = "${data.template_file.stack_name.rendered}"
  }
}

resource "aws_security_group" "kubernetes_ingress_controller_elb" {
  name   = "${data.template_file.stack_name.rendered}-k8s-ingress-controller-elb"
  vpc_id = "${data.terraform_remote_state.network.vpc_id}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-k8s-ingress-controller-elb"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"

    # Required for AWS cloud provider
    KubernetesCluster = "${data.template_file.stack_name.rendered}"
  }
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ssh_CI" {
  type                     = "ingress"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
  source_security_group_id = "${data.terraform_remote_state.hub_tools.jenkins_security_group_id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_bgp_kubernetes_master" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  source_security_group_id = "${aws_security_group.kubernetes_master.id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_bgp_kubernetes_worker" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_kubeproxy_kubernetes_master" {
  type                     = "ingress"
  from_port                = 10250
  to_port                  = 10250
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_master.id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_nodeport_elb" {
  type                     = "ingress"
  from_port                = 30000
  to_port                  = 32768
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_nodeport_elb.id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_nodeport_egress_kubernetes_nodeport_elb" {
  type                     = "egress"
  from_port                = 30000
  to_port                  = 32768
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.kubernetes_nodeport_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_elb_http" {
  type                     = "ingress"
  from_port                = 30080
  to_port                  = 30080
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_egress_kubernetes_worker_elb_http" {
  type                     = "egress"
  from_port                = 30080
  to_port                  = 30080
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_services_world_http" {
  type              = "ingress"
  from_port         = 80
  to_port           = 80
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_elb_https" {
  type                     = "ingress"
  from_port                = 30443
  to_port                  = 30443
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_egress_kubernetes_worker_elb_https" {
  type                     = "egress"
  from_port                = 30443
  to_port                  = 30443
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_services_world_https" {
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_elb_3000" {
  type                     = "ingress"
  from_port                = 30000
  to_port                  = 30000
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_egress_kubernetes_worker_elb_3000" {
  type                     = "egress"
  from_port                = 30000
  to_port                  = 30000
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_elb_https_alt" {
  type                     = "ingress"
  from_port                = 31443
  to_port                  = 31443
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_egress_kubernetes_worker_elb_https_alt" {
  type                     = "egress"
  from_port                = 31443
  to_port                  = 31443
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_services_world_https_alt" {
  type              = "ingress"
  from_port         = 8443
  to_port           = 8443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_elb_https_alt2" {
  type                     = "ingress"
  from_port                = 32443
  to_port                  = 32443
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_egress_kubernetes_worker_elb_https_alt2" {
  type                     = "egress"
  from_port                = 32443
  to_port                  = 32443
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_services_world_https_alt2" {
  type              = "ingress"
  from_port         = 9443
  to_port           = 9443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_elb_http_alt" {
  type                     = "ingress"
  from_port                = 31080
  to_port                  = 31080
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_egress_kubernetes_worker_elb_http_alt" {
  type                     = "egress"
  from_port                = 31080
  to_port                  = 31080
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_worker.id}"
  security_group_id        = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_services_world_http_alt" {
  type              = "ingress"
  from_port         = 8080
  to_port           = 8080
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_egress" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.kubernetes_worker.id}"
}
