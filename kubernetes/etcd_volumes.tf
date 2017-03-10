resource "aws_ebs_volume" "etcd" {
  count             = "${var.etcd_instance_count}"
  availability_zone = "${element(data.terraform_remote_state.network.availability_zones, count.index)}"
  size              = "${var.etcd_ebs_volume_size}"
  type              = "gp2"

  tags {
    Name        = "${data.template_file.stack_name.rendered}-etcd-${count.index+1}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
  }
}
