data "terraform_remote_state" "state" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "state_${var.environment}_${var.name}.tfstate"
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
