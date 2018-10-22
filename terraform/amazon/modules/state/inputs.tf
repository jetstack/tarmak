variable "name" {}

variable "project" {}

variable "contact" {}

variable "region" {}

variable "availability_zones" {
  type = "list"
}

variable "stack" {}

variable "public_zone" {}

variable "public_zone_id" {}

variable "state_bucket" {}

variable "stack_name_prefix" {}

variable "allowed_account_ids" {
  type = "list"
}

variable "environment" {}

variable "bucket_prefix" {}

data "template_file" "stack_name" {
  template = "${var.stack_name_prefix}${var.environment}-${var.name}"
}
