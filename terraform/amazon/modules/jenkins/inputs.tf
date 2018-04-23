variable "vpc_id" {}

variable "environment" {}

variable "project" {}

variable "contact" {}

variable "bastion_security_group_id" {}

variable "region" {}

variable "private_zone" {}

variable "private_zone_id" {}

variable "jenkins_ami" {}

variable "jenkins_instance_type" {}

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

variable "admin_ips" {
  type = "list"
}

variable "public_zone_id" {}

variable "stack_name_prefix" {}

variable "name" {}


