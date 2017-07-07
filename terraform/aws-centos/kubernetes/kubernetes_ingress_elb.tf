variable "ingress_elb_nodeport_http" {
  default = 32080
}

output "ingress_wildcard_fqdn" {
  value = "${aws_route53_record.ingress_wildcard.fqdn}"
}

#data "aws_acm_certificate" "wildcard" {
#  domain   = "*.${var.name}.${data.terraform_remote_state.hub_state.public_zone}"
#  statuses = ["ISSUED"]
#}

data "aws_elb_hosted_zone_id" "main" {}

resource "aws_route53_record" "ingress_wildcard" {
  zone_id = "${data.terraform_remote_state.hub_state.public_zone_id}"
  name    = "*.${var.name}"
  type    = "A"

  alias {
    name                   = "${aws_elb.ingress_controller.dns_name}"
    zone_id                = "${data.aws_elb_hosted_zone_id.main.id}"
    evaluate_target_health = true
  }
}

resource "aws_elb" "ingress_controller" {
  name         = "${format("%.20s-k8s-ingress", data.template_file.stack_name.rendered)}"
  subnets      = ["${data.terraform_remote_state.network.public_subnet_ids}"]
  idle_timeout = 600

  security_groups = [
    "${aws_security_group.kubernetes_ingress_controller_elb.id}",
  ]

  listener {
    instance_port     = "${var.ingress_elb_nodeport_http}"
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  #listener {
  #  instance_port      = "${var.ingress_elb_nodeport_http}"
  #  instance_protocol  = "http"
  #  lb_port            = 443
  #  lb_protocol        = "https"
  #  ssl_certificate_id = "${data.aws_acm_certificate.wildcard.arn}"
  #}

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 5
    timeout             = 3
    target              = "TCP:${var.ingress_elb_nodeport_http}"
    interval            = 10
  }
  tags {
    Name        = "${data.template_file.stack_name.rendered}-k8s-ingress"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
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

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_elb_http" {
  type                     = "ingress"
  from_port                = "${var.ingress_elb_nodeport_http}"
  to_port                  = "${var.ingress_elb_nodeport_http}"
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
  security_group_id        = "${aws_security_group.kubernetes_worker.id}"
}

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_elb_egress_http" {
  type                     = "egress"
  from_port                = "${var.ingress_elb_nodeport_http}"
  to_port                  = "${var.ingress_elb_nodeport_http}"
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

resource "aws_security_group_rule" "kubernetes_worker_allow_ingress_controller_services_world_https" {
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.kubernetes_ingress_controller_elb.id}"
}
