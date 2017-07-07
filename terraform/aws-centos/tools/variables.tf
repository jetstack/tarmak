variable "name" {}

variable "stack" {
  default = ""
}

variable "centos_ami" {
  type = "map"
}

variable "key_name" {}

variable "project" {}

variable "contact" {}

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

# TODO: restrict to admin IPs
variable "admin_ips" {
  type    = "list"
  default = ["0.0.0.0/0"]
}

variable "environment" {
  default = "nonprod"
}

variable "region" {
  default = "eu-west-1"
}

variable "state_context_name" {
  default = "hub"
}
