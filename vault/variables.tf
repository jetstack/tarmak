variable "name" {}

variable "stack" {
  default = ""
}

variable "centos_ami" {
  type = "map"
}

variable "project" {}

variable "contact" {}

variable "key_name" {}

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

variable "environment" {
  default = "nonprod"
}

variable "region" {
  default = "eu-west-1"
}

variable "consul_version" {
  default = "0.8.3"
}

variable "vault_version" {
  default = "0.7.2"
}

variable "vault_root_size" {
  default = "16"
}

variable "vault_data_size" {
  default = "10"
}

variable "instance_count" {
  default = 3
}

variable "vault_instance_type" {
  default = "t2.nano"
}

variable "consul_master_token" {}

variable "consul_encrypt" {}

data "template_file" "vault_unseal_key_name" {
  template = "vault-${var.environment}-unseal-key"
}

output "vault_kms_key_id" {
  value = "${element(split("/", data.terraform_remote_state.network.secrets_kms_arn), 1)}"
}

output "vault_unseal_key_name" {
  value = "${data.template_file.vault_unseal_key_name.rendered}"
}
