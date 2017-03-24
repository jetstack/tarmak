resource "aws_instance" "etcd" {
  depends_on           = ["aws_ebs_volume.etcd"]
  count                = "${var.etcd_instance_count}"
  availability_zone    = "${element(data.terraform_remote_state.network.availability_zones, count.index)}"
  ami                  = "${lookup(var.centos_ami, var.region)}"
  instance_type        = "${var.etcd_instance_type}"
  key_name             = "${var.key_name}"
  subnet_id            = "${element(data.terraform_remote_state.network.private_subnet_ids, count.index)}"
  iam_instance_profile = "${aws_iam_role.etcd.name}"
  monitoring           = true

  vpc_security_group_ids = [
    "${aws_security_group.etcd.id}",
  ]

  root_block_device {
    volume_type = "gp2"
    volume_size = "${var.etcd_root_volume_size}"
  }

  tags {
    Name               = "${data.template_file.stack_name.rendered}-k8s-etcd-${count.index+1}"
    Environment        = "${var.environment}"
    Project            = "${var.project}"
    Contact            = "${var.contact}"
    Etcd_Volume_Attach = "${data.template_file.stack_name.rendered}-k8s-etcd-${count.index+1}"
    Role               = "etcd"
    KubernetesCluster  = "${data.template_file.stack_name_dns.rendered}"
  }

  user_data = "${element(data.template_file.etcd_user_data.*.rendered, count.index)}"
}

data "template_file" "etcd_user_data" {
  count    = "${var.etcd_instance_count}"
  template = "${file("${path.module}/templates/puppet_agent_user_data.yaml")}"

  vars {
    puppet_fqdn        = "${data.terraform_remote_state.hub_tools.puppet_master_fqdn}"
    puppet_environment = "${replace(data.template_file.stack_name.rendered, "-", "_")}"
    puppet_runinterval = "${var.puppet_runinterval}"
    puppet_deploy_key  = "${data.terraform_remote_state.hub_tools.puppet_deploy_key}"

    vault_token = "${var.vault_init_token_etcd}"
    vault_ca    = "${base64encode(data.aws_s3_bucket_object.vault_ca.body)}"

    puppernetes_dns_root      = "${data.terraform_remote_state.hub_network.private_zones[0]}"
    puppernetes_role          = "etcd"
    puppernetes_hostname      = "etcd-${count.index+1}"
    puppernetes_cluster       = "${data.template_file.stack_name_dns.rendered}"
    puppernetes_environment   = "${var.environment}"
    puppernetes_desired_count = "${var.etcd_instance_count}"
    puppernetes_volume_id     = "${element(aws_ebs_volume.etcd.*.id, count.index)}"
  }
}

resource "aws_route53_record" "etcd" {
  zone_id = "${data.terraform_remote_state.hub_network.private_zone_ids[0]}"
  name    = "etcd-${count.index+1}.${data.template_file.stack_name_dns.rendered}"
  type    = "A"
  ttl     = "300"
  records = ["${element(aws_instance.etcd.*.private_ip, count.index)}"]
  count   = "${var.etcd_instance_count}"
}
