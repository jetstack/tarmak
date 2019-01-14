variable "vpc_id" {}

variable "environment" {}

variable "project" {}

variable "contact" {}

variable "bastion_security_group_id" {}

variable "region" {}

variable "private_zone" {}

variable "private_zone_id" {}

variable "secrets_bucket" {}

variable "public_subnet_ids" {
  type = "list"
}

variable "private_subnet_ids" {
  type = "list"
}

variable "key_name" {}

variable "jenkins_root_size" {}

variable "jenkins_ebs_size" {}

variable "certificate_arn" {}

variable "public_zone_id" {}

variable "stack_name_prefix" {}

variable "name" {}

variable "availability_zones" {
  type = "list"
}

variable "jenkins_admin_cidrs" {
  type = "list"
}

variable "tagging_control_policy_arn" {}
variable "wing_binary_read_policy_arn" {}
