# TODO: Enable as soon as ALBs are available (terraform 0.9+)
#resource "aws_alb" "main" {
#  name            = "${data.template_file.stack_name.rendered}"
#  subnets         = ["${data.terraform_remote_state.network.public_subnet_ids}"]
#  security_groups = ["${aws_security_group.lb_sg.id}"]
#
#  tags {
#    Name        = "${data.template_file.stack_name.rendered}"
#    Environment = "${var.environment}"
#    Project     = "${var.project}"
#    Contact     = "${var.contact}"
#  }
#}
#
#resource "aws_alb_listener" "front_end" {
#  load_balancer_arn = "${aws_alb.main.id}"
#  port              = "443"
#  protocol          = "HTTPS"
#
#  default_action {
#    target_group_arn = "${aws_alb_target_group.test.id}"
#    type             = "forward"
#  }
#}

resource "aws_security_group" "alb" {
  name        = "${data.template_file.stack_name.rendered}-alb"
  vpc_id      = "${data.terraform_remote_state.network.vpc_id}"
  description = "ALB for ${data.template_file.stack_name.rendered}"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-alb"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_security_group_rule" "alb_ingress_allow_https_admins" {
  type              = "ingress"
  protocol          = "tcp"
  from_port         = 443
  to_port           = 443
  cidr_blocks       = ["${var.admin_ips}"]
  security_group_id = "${aws_security_group.alb.id}"
}

resource "aws_security_group_rule" "alb_egress_allow_all" {
  type              = "egress"
  protocol          = -1
  from_port         = 0
  to_port           = 65535
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.alb.id}"
}

resource "aws_security_group_rule" "jenkins_ingress_allow_jenkins_alb" {
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 8080
  to_port                  = 8080
  source_security_group_id = "${aws_security_group.alb.id}"
  security_group_id        = "${aws_security_group.jenkins.id}"
}
