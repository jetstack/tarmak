variable "name" {}

variable "project" {}

variable "contact" {}

variable "key_name" {}

variable "region" {}

variable "stack" {}

variable "state_bucket" {}

variable "stack_name_prefix" {}

variable "allowed_account_ids" {
  type = "list"
}

variable "environment" {}

variable "state_cluster_name" {}

variable "vault_cluster_name" {}

variable "tools_cluster_name" {}

# data.terraform_remote_state.hub_state.secrets_bucket.0
variable "secrets_bucket" {}

# data.terraform_remote_state.network.private_subnet_ids
variable "private_subnet_ids" {
  type = "list"
}

# data.terraform_remote_state.network.public_subnet_ids
variable "public_subnet_ids" {
  type = "list"
}

data "template_file" "stack_name" {
  template = "${var.stack_name_prefix}${var.environment}-${var.name}"
}

variable "internal_fqdns" {
  type = "list"
}

variable "vault_kms_key_id" {}

variable "vault_unseal_key_name" {}

# template variables
variable "availability_zones" {
  type = "list"
}

variable "api_admin_cidrs" {
  type = "list"
}

variable "vpc_id" {}

variable "private_zone_id" {}

variable "vault_ca" {}

variable "vault_url" {}

variable "private_zone" {}

variable "public_zone" {}

variable "public_zone_id" {}

variable "vault_security_group_id" {}

variable "bastion_security_group_id" {}

variable "elb_access_logs_public_enabled" {}
variable "elb_access_logs_public_bucket" {}
variable "elb_access_logs_public_bucket_prefix" {}
variable "elb_access_logs_public_bucket_interval" {}
variable "elb_access_logs_internal_enabled" {}
variable "elb_access_logs_internal_bucket" {}
variable "elb_access_logs_internal_bucket_prefix" {}
variable "elb_access_logs_internal_bucket_interval" {}
