variable "network" {}

variable "name" {}

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
  template = "${var.stack_name_prefix}${var.environment}_${var.name}"
}

variable "allowed_account_ids" {
  type    = "list"
  default = ["513013539150"]
}

variable "state_buckets" {
  type    = "list"
  default = []
}

variable "public_zones" {
  type    = "list"
  default = []
}

variable "private_zones" {
  type    = "list"
  default = []
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

variable "project" {
  default = "cynosura"
}

variable "contact" {
  default = "matt.turner@skyscanner.net"
}

variable "bucket_prefix" {
  default = ""
}
