# data.terraform_remote_state.state.public_zone
variable "public_zone" {}

variable "environment" {}
variable "stack_name_prefix" {}
variable "name" {}

# data.terraform_remote_state.network.vpc_id
variable "vpc_id" {}

variable "project" {}
variable "contact" {}
variable "bastion_ami" {}
variable "bastion_instance_type" {}
variable "bastion_min_instance_count" {}

# data.terraform_remote_state.network.public_subnet_ids
variable "public_subnet_ids" {
  type = "list"
}

variable "key_name" {}
variable "bastion_root_size" {}

# TODO: restrict to admin IPs
variable "bastion_admin_cidrs" {
  type = "list"
}

# data.terraform_remote_state.state.public_zone_id
variable "public_zone_id" {}

# data.terraform_remote_state.network.private_zone_id.0
variable "private_zone_id" {}

variable "bastion_iam_additional_policy_arns" {
  type = "list"
}
