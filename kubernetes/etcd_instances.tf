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
    "${data.terraform_remote_state.hub_tools.remote_admin_security_group_id}",
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
    Etcd_Volume_Attach = "kubernetes.${data.template_file.stack_name.rendered}.etcd.${count.index}"
    Role               = "etcd"
  }

  user_data = "${data.template_file.etcd_user_data.rendered}"
}

data "template_file" "etcd_user_data" {
  template = "${file("${path.module}/templates/puppet_agent_user_data.yaml")}"

  vars {
    puppet_fqdn        = "puppetmaster.${data.terraform_remote_state.hub_network.private_zone_ids[0]}"
    puppet_environment = "${replace(data.template_file.stack_name.rendered, "-", "_")}"
    puppet_runinterval = "${var.puppet_runinterval}"
    vault_token        = "${var.vault_init_token_etcd}"
    dns_root           = "${data.terraform_remote_state.hub_network.private_zone_ids[0]}"
  }
}

resource "aws_route53_record" "etcd" {
  zone_id = "${data.terraform_remote_state.hub_network.private_zone_ids[0]}"
  name    = "etcd-${count.index}.${data.template_file.stack_name.rendered}"
  type    = "A"
  ttl     = "300"
  records = ["${element(aws_instance.etcd.*.private_ip, count.index)}"]
  count   = "${var.etcd_instance_count}"
}
