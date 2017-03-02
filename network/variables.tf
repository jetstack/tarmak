variable "network" {
  default = "10.61.0.0/20"
}

variable "environment" {
  default = "non_prod"
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

variable "vpc_name" {}
