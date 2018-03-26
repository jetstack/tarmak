# network
region = "eu-west-1"
tools_cluster_name = "cluster"
availability_zones = ["eu-west-1b"]
name = "cluster"
contact = "luke.addison@jetstack.io"
key_name = "tarmak_luke"
public_zone_id = "Z3GXMBTQUCZJRP"
private_zone = "tarmak.local"
state_cluster_name = "cluster"
vault_cluster_name = "cluster"
environment = "luke"
network = "10.99.0.0/16"
project = "luke"
bucket_prefix = "aws-tarmak-luke-"
public_zone = "develop.tarmak.org"
state_bucket = "aws-tarmak-luke-eu-west-1-terraform-state"

# tools
bastion_ami = "ami-13f9ad6a"
tools_cluster_name = "cluster"

# vault
vault_ami = "ami-13f9ad6a"
instance_count = 3

# kubernetes
kubernetes_etcd_ami = "ami-13f9ad6a"
kubernetes_worker_ami = "ami-13f9ad6a"
kubernetes_master_ami = "ami-13f9ad6a"
secrets_bucket = "aws-tarmak-luke-luke-eu-west-1-secrets"