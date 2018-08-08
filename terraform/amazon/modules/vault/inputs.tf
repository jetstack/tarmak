variable "name" {}

variable "stack" {}

variable "project" {}

variable "contact" {}

variable "key_name" {}

variable "region" {}

variable "vault_ami" {}

variable "state_bucket" {}

variable "stack_name_prefix" {}

variable "allowed_account_ids" {
  type    = "list"
}

variable "environment" {}

variable "consul_version" {}

variable "vault_version" {}

variable "vault_root_size" {}

variable "vault_data_size" {}

variable "vault_min_instance_count" {}

variable "vault_instance_type" {}

variable "state_cluster_name" {}

# data.terraform_remote_state.network.private_zone.0
variable "private_zone" {}

# data.terraform_remote_state.network.private_zone_id.0
variable "private_zone_id" {}

# data.terraform_remote_state.state.secrets_bucket.0
variable "secrets_bucket" {}

# data.terraform_remote_state.state.secrets_kms_arn.0
variable "secrets_kms_arn" {}

# data.terraform_remote_state.state.backups_bucket.0
variable "backups_bucket" {}

# data.terraform_remote_state.network.private_subnet_ids
variable "private_subnet_ids" {
  type = "list"
}

# data.terraform_remote_state.network.private_subnets
variable "private_subnets" {
  type = "list"
}

# data.terraform_remote_state.network.availability_zones
variable "availability_zones" {
  type = "list"
}

# data.terraform_remote_state.tools.bastion_security_group_id
variable "bastion_security_group_id" {}

# data.terraform_remote_state.network.vpc_id
variable "vpc_id" {}

variable "vault_cluster_name" {}

data "template_file" "stack_name" {
  template = "${var.stack_name_prefix}${var.environment}-${var.name}"
}

data "template_file" "vault_unseal_key_name" {
  template = "vault-${var.environment}-"
}

variable "bastion_status" {}
