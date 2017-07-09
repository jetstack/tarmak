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

variable "public_zone" {}

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

variable "bucket_prefix" {
  default = ""
}
