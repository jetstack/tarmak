variable "network" {}

variable "name" {}

variable "project" {}

variable "contact" {}

variable "region" {}

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

variable "state_context_name" {
  default = "hub"
}
