data "terraform_remote_state" "hub_state" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "state_${var.environment}_hub.tfstate"
  }
}

data "terraform_remote_state" "hub_network" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "network_${var.environment}_hub.tfstate"
  }
}

data "terraform_remote_state" "hub_tools" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "tools_${var.environment}_hub.tfstate"
  }
}

data "terraform_remote_state" "hub_vault" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "vault_${var.environment}_hub.tfstate"
  }
}

data "terraform_remote_state" "network" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "network_${var.environment}_${var.name}.tfstate"
  }
}
