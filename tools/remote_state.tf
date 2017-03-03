data "terraform_remote_state" "network" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "network_${var.environment}_${var.name}.tfstate"
  }
}
