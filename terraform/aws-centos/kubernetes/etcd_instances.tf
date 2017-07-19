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
    KubernetesCluster  = "${data.template_file.stack_name.rendered}"
    tarmak_role        = "etcd-${count.index+1}"
  }

  user_data = "${element(data.template_file.etcd_user_data.*.rendered, count.index)}"

  lifecycle {
    ignore_changes = ["volume_tags"]
  }
}

data "template_file" "etcd_user_data" {
  count    = "${var.etcd_instance_count}"
  template = "${file("${path.module}/templates/puppet_agent_user_data.yaml")}"

  vars {
    vault_token = "${var.vault_init_token_etcd}"
    vault_ca    = "${base64encode(data.terraform_remote_state.hub_vault.vault_ca)}"
    vault_url   = "${data.terraform_remote_state.hub_vault.vault_url}"

    tarmak_dns_root      = "${data.terraform_remote_state.hub_network.private_zone}"
    tarmak_role          = "etcd"
    tarmak_hostname      = "etcd-${count.index+1}"
    tarmak_cluster       = "${data.template_file.stack_name.rendered}"
    tarmak_environment   = "${var.environment}"
    tarmak_desired_count = "${var.etcd_instance_count}"
    tarmak_volume_id     = "${element(aws_ebs_volume.etcd.*.id, count.index)}"
  }
}

resource "aws_route53_record" "etcd" {
  zone_id = "${data.terraform_remote_state.hub_network.private_zone_id}"
  name    = "etcd-${count.index+1}.${data.template_file.stack_name.rendered}"
  type    = "A"
  ttl     = "300"
  records = ["${element(aws_instance.etcd.*.private_ip, count.index)}"]
  count   = "${var.etcd_instance_count}"
}
