# tools
/*data "terraform_remote_state" "state" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "${var.environment}/${var.state_cluster_name}/state.tfstate"
  }
}

data "terraform_remote_state" "network" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "${var.environment}/${var.name}/network.tfstate"
  }
}

# network
data "terraform_remote_state" "vpc_peer_stack" {
  count   = "${signum(length(var.vpc_peer_stack))}"
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "network_${var.environment}_${var.vpc_peer_stack}.tfstate"
  }
}*/

/*
I don't think this is used at all in the network stack
data "terraform_remote_state" "hub_state" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "${var.environment}/${var.state_cluster_name}/state.tfstate"
  }
}
*/