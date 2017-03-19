variable "name" {}

variable "stack" {
  default = ""
}

variable "centos_ami" {
  default = {
    eu-west-1 = "ami-d2cbfab4"
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
  template = "${var.stack_name_prefix}${var.environment}_${var.name}"
}

data "template_file" "stack_name_dns" {
  template = "${var.stack_name_prefix}${var.environment}-${var.name}"
}

variable "allowed_account_ids" {
  type    = "list"
  default = []
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
  default = "m3.medium"
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

variable "kubernetes_worker_spot_price" {
  default = ""
}
