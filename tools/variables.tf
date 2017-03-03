variable "name" {}

variable "stack" {
  default = ""
}

variable "centos_ami" {
  default = {
    eu-west-1 = "ami-7abd0209"
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
