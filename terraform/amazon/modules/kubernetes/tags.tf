resource "awstag_ec2_tag" "public_subnet" {
  count  = "${length(var.availability_zones)}"
  ec2_id = "${element(var.public_subnet_ids, count.index)}"
  key    = "kubernetes.io/cluster/${data.template_file.stack_name.rendered}"
  value  = ""
}

resource "awstag_ec2_tag" "private_subnet" {
  count  = "${length(var.availability_zones)}"
  ec2_id = "${element(var.private_subnet_ids, count.index)}"
  key    = "kubernetes.io/cluster/${data.template_file.stack_name.rendered}"
  value  = ""
}
