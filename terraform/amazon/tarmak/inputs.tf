variable "network" {}

variable "name" {}

variable "project" {}

variable "contact" {}

variable "region" {}

# data.terraform_remote_state.vpc_peer_stack.vpc_id
variable "peer_vpc_id" {
  default = ""
}

variable "availability_zones" {
  type = "list"
}

variable "stack" {
  default = ""
}

variable "state_bucket" {
  default = ""
}

variable "stack_name_prefix" {
  default = ""
}

data "template_file" "stack_name" {
  template = "${var.stack_name_prefix}${var.environment}-${var.name}"
}

variable "allowed_account_ids" {
  type    = "list"
  default = []
}

variable "vpc_peer_stack" {
  default = ""
}

variable "environment" {
  default = "nonprod"
}

variable "private_zone" {
  default = ""
}

variable "state_cluster_name" {
  default = "hub"
}

# data.terraform_remote_state.vpc_peer_stack.vpc_net
variable "vpc_net" {
  default = ""
}

# data.terraform_remote_state.vpc_peer_stack.route_table_public_ids
variable "route_table_public_ids" {
  type = "list"
  default = []
}

# data.terraform_remote_state.vpc_peer_stack.route_table_private_ids
variable "route_table_private_ids" {
  type = "list"
  default = []
}

# data.terraform_remote_state.vpc_peer_stack.private_zone_id
variable "private_zone_id" {
  default = ""
}

# tools
variable "bastion_ami" {}
variable "bastion_instance_type" {
  default = "t2.nano"
}
variable "bastion_root_size" {
  default = "16"
}
# TODO: restrict to admin IPs
variable "admin_ips" {
  type    = "list"
  default = ["0.0.0.0/0"]
}
variable "key_name" {}
variable "public_zone" {}
variable "public_zone_id" {}

# vault
variable "consul_version" {
  default = "1.0.6"
}
variable "vault_version" {
  default = "0.9.5"
}
variable "vault_root_size" {
  default = "16"
}
variable "vault_data_size" {
  default = "10"
}
variable "instance_count" {}
variable "vault_instance_type" {
  default = "t2.nano"
}
variable "vault_ami" {}

# state 
variable "bucket_prefix" {}