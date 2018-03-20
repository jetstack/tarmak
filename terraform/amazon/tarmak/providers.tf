/*provider "tarmak" {
  socket_path = "/tmp/tarmak-connector.sock"
}*/

provider "template" {
  version = "~> 1.0"
}

provider "aws" {
  region              = "${var.region}"
  allowed_account_ids = ["${var.allowed_account_ids}"]
  version             = "~> 1.7.1"
}