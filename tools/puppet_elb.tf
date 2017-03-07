resource "aws_elb" "puppet_master" {
  name            = "${replace(data.template_file.stack_name.rendered,"_","-")}-puppet-master"
  subnets         = ["${data.terraform_remote_state.network.public_subnet_ids}"]
  security_groups = ["${aws_security_group.puppet_master_elb.id}"]

  listener {
    instance_port      = 443
    instance_protocol  = "https"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = "${data.aws_acm_certificate.wildcard.arn}"
  }

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 3
    target              = "HTTPS:443/users/login"
    interval            = 30
  }

  cross_zone_load_balancing = true
  idle_timeout              = 600

  tags {
    Name        = "${data.template_file.stack_name.rendered}-puppet_master"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_elb_attachment" "puppet_master" {
  elb      = "${aws_elb.puppet_master.name}"
  instance = "${aws_instance.puppet_master.id}"
}

resource "aws_security_group" "puppet_master_elb" {
  name        = "${data.template_file.stack_name.rendered}-puppet_master-elb"
  vpc_id      = "${data.terraform_remote_state.network.vpc_id}"
  description = "ELB for ${data.template_file.stack_name.rendered}-puppet_master"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-puppet_master-elb"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_security_group_rule" "puppet_master_elb_ingress_allow_admins" {
  type              = "ingress"
  protocol          = "tcp"
  from_port         = 443
  to_port           = 443
  cidr_blocks       = ["${var.admin_ips}"]
  security_group_id = "${aws_security_group.puppet_master_elb.id}"
}

resource "aws_security_group_rule" "puppet_master_elb_egress_allow_all" {
  type              = "egress"
  protocol          = "tcp"
  from_port         = 0
  to_port           = 65535
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.puppet_master_elb.id}"
}

resource "aws_security_group_rule" "puppet_master_ingress_allow_foreman_elb" {
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 443
  to_port                  = 443
  source_security_group_id = "${aws_security_group.puppet_master_elb.id}"
  security_group_id        = "${aws_security_group.puppet_master.id}"
}

resource "aws_route53_record" "foreman_elb" {
  zone_id = "${data.terraform_remote_state.network.public_zone_ids[0]}"
  name    = "foreman"
  type    = "A"

  alias {
    name                   = "${aws_elb.puppet_master.dns_name}"
    zone_id                = "${data.aws_elb_hosted_zone_id.main.id}"
    evaluate_target_health = true
  }
}

output "foreman_url" {
  value = "https://${aws_route53_record.foreman_elb.fqdn}"
}
