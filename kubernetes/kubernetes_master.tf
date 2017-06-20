resource "aws_elb" "kubernetes_master" {
  name         = "${data.template_file.stack_name.rendered}-master"
  subnets      = ["${data.terraform_remote_state.network.private_subnet_ids}"]
  idle_timeout = 3600
  internal     = true

  security_groups = [
    "${aws_security_group.kubernetes_master_elb.id}",
  ]

  listener {
    instance_port     = 6443
    instance_protocol = "tcp"
    lb_port           = 6443
    lb_protocol       = "tcp"
  }

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 3
    target              = "TCP:6443"
    interval            = 30
  }

  tags {
    Name        = "${data.template_file.stack_name.rendered}-master-elb"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}

resource "aws_launch_configuration" "kubernetes_master" {
  lifecycle {
    create_before_destroy = true
  }

  image_id             = "${lookup(var.centos_ami, var.region)}"
  instance_type        = "${var.kubernetes_master_instance_type}"
  name_prefix          = "${data.template_file.stack_name.rendered}-k8s-master-"
  key_name             = "${var.key_name}"
  iam_instance_profile = "${aws_iam_role.kubernetes_master.name}"

  security_groups = [
    "${aws_security_group.kubernetes_master.id}",
  ]

  root_block_device {
    volume_type = "gp2"
    volume_size = "${var.kubernetes_master_root_volume_size}"
  }

  user_data = "${data.template_file.kubernetes_master_user_data.rendered}"
}

data "template_file" "kubernetes_master_user_data" {
  template = "${file("${path.module}/templates/puppet_agent_user_data.yaml")}"

  vars {
    puppet_fqdn        = "${data.terraform_remote_state.hub_tools.puppet_master_fqdn}"
    puppet_environment = "${replace(data.template_file.stack_name.rendered, "-", "_")}"
    puppet_runinterval = "${var.puppet_runinterval}"
    puppet_deploy_key  = "${data.terraform_remote_state.hub_tools.puppet_deploy_key}"

    vault_token = "${var.vault_init_token_master}"
    vault_ca    = "${base64encode(data.aws_s3_bucket_object.vault_ca.body)}"

    tarmak_dns_root      = "${data.terraform_remote_state.hub_network.private_zone}"
    tarmak_role          = "master"
    tarmak_hostname      = "master"
    tarmak_cluster       = "${data.template_file.stack_name.rendered}"
    tarmak_environment   = "${var.environment}"
    tarmak_desired_count = "${var.kubernetes_master_count}"
    tarmak_volume_id     = ""
  }
}

resource "aws_autoscaling_group" "kubernetes_master" {
  name                      = "kubernetes-${data.template_file.stack_name.rendered}-master"
  max_size                  = "${var.kubernetes_master_count}"
  min_size                  = "${var.kubernetes_master_count}"
  health_check_grace_period = 600
  health_check_type         = "EC2"
  desired_capacity          = "${var.kubernetes_master_count}"
  vpc_zone_identifier       = ["${data.terraform_remote_state.network.private_subnet_ids}"]
  launch_configuration      = "${aws_launch_configuration.kubernetes_master.name}"
  load_balancers            = ["${aws_elb.kubernetes_master.name}"]

  tag {
    key                 = "Name"
    value               = "${data.template_file.stack_name.rendered}-k8s-master"
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
    value               = "master"
    propagate_at_launch = true
  }

  # Required for AWS cloud provider
  tag {
    key                 = "KubernetesCluster"
    value               = "${data.template_file.stack_name.rendered}"
    propagate_at_launch = true
  }
}

resource "aws_route53_record" "kubernetes_master" {
  zone_id = "${data.terraform_remote_state.hub_network.private_zone_id}"
  name    = "api.${data.template_file.stack_name.rendered}"
  type    = "CNAME"
  ttl     = "60"
  records = ["${aws_elb.kubernetes_master.dns_name}"]
}
