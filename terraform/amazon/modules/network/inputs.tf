variable "network" {}

variable "name" {}

variable "project" {}

variable "contact" {}

variable "region" {}

# data.terraform_remote_state.vpc_peer_stack.vpc_id
variable "peer_vpc_id" {}

variable "availability_zones" {
  type = "list"
}

variable "stack" {}

variable "state_bucket" {}

variable "stack_name_prefix" {}

variable "allowed_account_ids" {
  type = "list"
}

variable "vpc_peer_stack" {}

variable "environment" {}

variable "private_zone" {}

variable "state_cluster_name" {}

# data.terraform_remote_state.vpc_peer_stack.vpc_net
variable "vpc_net" {}

# data.terraform_remote_state.vpc_peer_stack.route_table_public_ids
variable "route_table_public_ids" {
  type = "list"
}

# data.terraform_remote_state.vpc_peer_stack.route_table_private_ids
variable "route_table_private_ids" {
  type = "list"
}

# data.terraform_remote_state.vpc_peer_stack.private_zone_id
variable "private_zone_id" {}

data "template_file" "stack_name" {
  template = "${var.stack_name_prefix}${var.environment}-${var.name}"
}
