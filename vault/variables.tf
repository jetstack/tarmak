variable "name" {}

variable "stack" {
  default = ""
}

variable "coreos_ami" {
  default = {
    eu-west-1 = "ami-4829072e"
  }
}

variable "key_name" {
  default = "skyscanner_non_prod"
}

variable "state_bucket" {
  default = ""
}

variable "stack_name_prefix" {
  default = ""
}

data "template_file" "stack_name" {
  template = "${var.stack_name_prefix}${var.environment}_${var.name}"
}

variable "allowed_account_ids" {
  type    = "list"
  default = ["513013539150"]
}

variable "environment" {
  default = "nonprod"
}

variable "region" {
  default = "eu-west-1"
}

variable "project" {
  default = "cynosura"
}

variable "contact" {
  default = "matt.turner@skyscanner.net"
}

variable "consul_version" {
  default = "0.7.5"
}

variable "vault_version" {
  default = "0.6.5"
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
