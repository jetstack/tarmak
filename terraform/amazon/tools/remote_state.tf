data "terraform_remote_state" "state" {
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
