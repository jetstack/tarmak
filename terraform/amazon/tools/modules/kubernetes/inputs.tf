variable "name" {}

variable "project" {}

variable "contact" {}

variable "key_name" {}

variable "region" {}

variable "stack" {}

variable "state_bucket" {}

variable "stack_name_prefix" {}

variable "allowed_account_ids" {
  type    = "list"
}

variable "environment" {}

variable "vault_init_token_master" {}

variable "vault_init_token_worker" {}

variable "vault_init_token_etcd" {}

variable "state_cluster_name" {}

variable "vault_cluster_name" {}

variable "tools_cluster_name" {}

# data.terraform_remote_state.hub_state.secrets_bucket.0
variable "secrets_bucket" {}

# data.terraform_remote_state.network.private_subnet_ids
variable "private_subnet_ids" {
  type    = "list"
}

# data.terraform_remote_state.network.public_subnet_ids
variable "public_subnet_ids" {
  type    = "list"
}

data "template_file" "stack_name" {
  template = "${var.stack_name_prefix}${var.environment}-${var.name}"
}