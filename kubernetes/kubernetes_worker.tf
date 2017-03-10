resource "aws_elb" "kubernetes_nodeport" {
  name     = "${data.template_file.stack_name_dns.rendered}-k8s-nodeport"
  subnets  = ["${data.terraform_remote_state.network.private_subnet_ids}"]
  internal = true

  security_groups = [
    "${aws_security_group.kubernetes_nodeport_elb.id}",
  ]

  listener {
    instance_port     = 31131
    instance_protocol = "tcp"
    lb_port           = 31131
    lb_protocol       = "tcp"
  }

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 3
    target              = "TCP:31131"
    interval            = 30
  }

  tags {
    Name        = "${data.template_file.stack_name.rendered}-k8s-nodeport-elb"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_elb" "kubernetes_ingress_controller" {
  name     = "${data.template_file.stack_name_dns.rendered}-k8s-ingress"
  subnets = ["${data.terraform_remote_state.network.public_subnet_ids}"]

  security_groups = [
    "${aws_security_group.kubernetes_ingress_controller_elb.id}",
  ]

  listener {
    instance_port     = 30080
    instance_protocol = "tcp"
    lb_port           = 80
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 30443
    instance_protocol = "tcp"
    lb_port           = 443
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 30000
    instance_protocol = "tcp"
    lb_port           = 3000
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 31080
    instance_protocol = "http"
    lb_port           = 8080
    lb_protocol       = "http"
  }

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 3
    target              = "TCP:30080"
    interval            = 30
  }

  tags {
    Name        = "${data.template_file.stack_name.rendered}-k8s-ingress-controller-elb"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_proxy_protocol_policy" "ingress" {
  load_balancer  = "${aws_elb.kubernetes_ingress_controller.name}"
  instance_ports = ["30080", "30443", "30000"]
}

resource "aws_launch_configuration" "kubernetes_worker" {
  lifecycle {
    create_before_destroy = true
  }

  image_id             = "${lookup(var.centos_ami, var.region)}"
  instance_type        = "${var.kubernetes_worker_instance_type}"
  name_prefix          = "${data.template_file.stack_name.rendered}-k8s-worker-"
  key_name             = "${var.key_name}"
  iam_instance_profile = "${aws_iam_role.kubernetes_worker.name}"

  security_groups = [
    "${aws_security_group.kubernetes_worker.id}",
    "${data.terraform_remote_state.hub_tools.remote_admin_security_group_id}",
  ]

  root_block_device {
    volume_type = "gp2"
    volume_size = "${var.kubernetes_worker_root_volume_size}"
  }

  ebs_block_device {
    device_name = "/dev/sdd"
    volume_size = "${var.kubernetes_worker_docker_volume_size}"
    volume_type = "gp2"
  }

  user_data = "${data.template_file.kubernetes_worker_user_data.rendered}"
}

data "template_file" "kubernetes_worker_user_data" {
  template = "${file("${path.module}/templates/puppet_agent_user_data.yaml")}"

  vars {
    puppet_fqdn        = "puppetmaster.${data.terraform_remote_state.hub_network.private_zone_ids[0]}"
    puppet_environment = "${replace(data.template_file.stack_name.rendered, "-", "_")}"
    puppet_runinterval = "${var.puppet_runinterval}"
    vault_token        = "${var.vault_init_token_worker}"
    dns_root           = "${data.terraform_remote_state.hub_network.private_zone_ids[0]}"
  }
}

resource "aws_autoscaling_group" "kubernetes_worker" {
  name                      = "kubernetes-${data.template_file.stack_name.rendered}-worker"
  max_size                  = 300
  min_size                  = "${var.kubernetes_worker_count}"
  health_check_grace_period = 600
  health_check_type         = "EC2"
  vpc_zone_identifier       = ["${data.terraform_remote_state.network.private_subnet_ids}"]
  launch_configuration      = "${aws_launch_configuration.kubernetes_worker.name}"
  load_balancers            = ["${aws_elb.kubernetes_nodeport.name}", "${aws_elb.kubernetes_ingress_controller.name}"]

  tag {
    key                 = "Name"
    value               = "${data.template_file.stack_name.rendered}-k8s-worker"
    propagate_at_launch = true
  }

  tag {
    key                 = "Environment"
    value               = "${var.environment}"
    propagate_at_launch = true
  }

  tag {
    key                 = "Project"
    value               = "${var.project}"
    propagate_at_launch = true
  }

  tag {
    key                 = "Contact"
    value               = "${var.contact}"
    propagate_at_launch = true
  }

  tag {
    key                 = "Role"
    value               = "worker"
    propagate_at_launch = true
  }

  # Required for AWS cloud provider
  tag {
    key                 = "KubernetesCluster"
    value               = "${data.template_file.stack_name.rendered}"
    propagate_at_launch = true
  }
}

resource "aws_route53_record" "kubernetes_nodeport" {
  zone_id = "${data.terraform_remote_state.hub_network.private_zone_ids[0]}"
  name    = "kube-nodes.${replace(data.template_file.stack_name.rendered,"_","-")}"
  type    = "CNAME"
  ttl     = "60"
  records = ["${aws_elb.kubernetes_nodeport.dns_name}"]
}
