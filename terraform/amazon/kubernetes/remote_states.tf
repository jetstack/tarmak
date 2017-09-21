data "terraform_remote_state" "hub_state" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "${var.environment}/${var.state_cluster_name}/state.tfstate"
  }
}

data "terraform_remote_state" "hub_network" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "${var.environment}/${var.tools_cluster_name}/network.tfstate"
  }
}

data "terraform_remote_state" "hub_tools" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "${var.environment}/${var.tools_cluster_name}/tools.tfstate"
  }
}

data "terraform_remote_state" "hub_vault" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "${var.environment}/${var.vault_cluster_name}/vault.tfstate"
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
