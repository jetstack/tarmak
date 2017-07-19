resource "aws_launch_configuration" "kubernetes_worker" {
  lifecycle {
    create_before_destroy = true
  }

  image_id             = "${lookup(var.centos_ami, var.region)}"
  instance_type        = "${var.kubernetes_worker_instance_type}"
  name_prefix          = "${data.template_file.stack_name.rendered}-k8s-worker-"
  key_name             = "${var.key_name}"
  iam_instance_profile = "${aws_iam_role.kubernetes_worker.name}"

  spot_price = "${var.kubernetes_worker_spot_price}"

  security_groups = [
    "${aws_security_group.kubernetes_worker.id}",
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
    region = "${var.region}"

    puppet_tar_gz_bucket_path = "${data.terraform_remote_state.hub_state.secrets_bucket}/${aws_s3_bucket_object.puppet-tar-gz.key}"

    vault_token = "${var.vault_init_token_worker}"
    vault_ca    = "${base64encode(data.terraform_remote_state.hub_vault.vault_ca)}"
    vault_url   = "${data.terraform_remote_state.hub_vault.vault_url}"

    tarmak_dns_root      = "${data.terraform_remote_state.hub_network.private_zone}"
    tarmak_role          = "worker"
    tarmak_hostname      = "worker"
    tarmak_cluster       = "${data.template_file.stack_name.rendered}"
    tarmak_environment   = "${var.environment}"
    tarmak_desired_count = 0
    tarmak_volume_id     = ""
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
  load_balancers            = ["${aws_elb.ingress_controller.name}"]

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

  tag {
    key                 = "tarmak_role"
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
