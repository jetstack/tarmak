variable "name" {}

variable "stack" {
  default = ""
}

variable "centos_ami" {
  default = {
    eu-west-1 = "ami-b790a3d1"
  }
}

variable "key_name" {
  default = "jetstack_nonprod"
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

variable "project" {
  default = "p9s"
}

variable "contact" {
  default = "christian@jetstack.io"
}
