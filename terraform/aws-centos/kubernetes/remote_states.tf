data "terraform_remote_state" "hub_state" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "${var.environment}/${var.state_context_name}/state.tfstate"
  }
}

data "terraform_remote_state" "hub_network" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "${var.environment}/${var.tools_context_name}/network.tfstate"
  }
}

data "terraform_remote_state" "hub_tools" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "${var.environment}/${var.tools_context_name}/tools.tfstate"
  }
}

data "terraform_remote_state" "hub_vault" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "${var.environment}/${var.vault_context_name}/vault.tfstate"
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
