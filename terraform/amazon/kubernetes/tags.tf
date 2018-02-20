resource "awstag_ec2_tag" "public_subnet" {
  count = "${length(data.terraform_remote_state.network.public_subnet_ids)}"
  ec2_id = "${element(data.terraform_remote_state.network.public_subnet_ids, count.index)}"
  key = "kubernetes.io/cluster/${data.template_file.stack_name.rendered}"
  value = ""
}

resource "awstag_ec2_tag" "private_subnet" {
  count = "${length(data.terraform_remote_state.network.private_subnet_ids)}"
  ec2_id = "${element(data.terraform_remote_state.network.private_subnet_ids, count.index)}"
  key = "kubernetes.io/cluster/${data.template_file.stack_name.rendered}"
  value = ""
}