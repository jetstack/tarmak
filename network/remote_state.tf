data "terraform_remote_state" "hub_state" {
  backend = "s3"

  config {
    region = "${var.region}"
    bucket = "${var.state_bucket}"
    key    = "state_${var.environment}_hub.tfstate"
  }
}
