variable "network" {}

variable "name" {}

variable "project" {}

variable "contact" {}

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

variable "region" {
  default = "eu-west-1"
}

variable "availability_zones" {
  default = ["eu-west-1a", "eu-west-1b", "eu-west-1c"]
}

variable "private_zone" {
  default = ""
}
