variable "name" {}

variable "project" {}

variable "contact" {}

variable "key_name" {}

variable "region" {}

variable "centos_ami" {
  type = "map"
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

variable "environment" {
  default = "nonprod"
}

variable "vault_init_token_master" {
  default = ""
}

variable "vault_init_token_worker" {
  default = ""
}

variable "vault_init_token_etcd" {
  default = ""
}

variable "puppet_runinterval" {
  default = "10m"
}

# etcd
variable "etcd_ebs_volume_size" {
  default = 20
}

variable "kubernetes_worker_docker_volume_size" {
  default = 50
}

variable "state_context_name" {
  default = "hub"
}

variable "vault_context_name" {
  default = "hub"
}

variable "tools_context_name" {
  default = "hub"
}
