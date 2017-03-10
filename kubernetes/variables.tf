variable "name" {}

variable "stack" {
  default = ""
}

variable "centos_ami" {
  default = {
    eu-west-1 = "ami-76c2e910"
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

data "template_file" "stack_name_dns" {
  template = "${var.stack_name_prefix}${var.environment}-${var.name}"
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

variable "vault_init_token_master" {
  default = ""
}

variable "vault_init_token_worker" {
  default = ""
}

variable "vault_init_token_etcd" {
  default = ""
}

variable "puppet_runinterval" {
  default = "10m"
}

# etcd
variable "etcd_instance_type" {
  default = "c4.large"
}

variable "etcd_instance_count" {
  default = "5"
}

variable "etcd_root_volume_size" {
  default = 32
}

variable "etcd_ebs_volume_size" {
  default = 20
}

# kubernetes master
variable "kubernetes_master_instance_type" {
  default = "c4.large"
}

variable "kubernetes_master_count" {
  default = 3
}

variable "kubernetes_master_root_volume_size" {
  default = 32
}

# kebernetes worker
variable "kubernetes_worker_instance_type" {
  default = "c4.large"
}

variable "kubernetes_worker_count" {
  default = 3
}

variable "kubernetes_worker_root_volume_size" {
  default = 32
}

variable "kubernetes_worker_docker_volume_size" {
  default = 50
}
