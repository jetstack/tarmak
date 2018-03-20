provider "tarmak" {
  socket_path = "/tmp/tarmak-connector.sock"
}

provider "template" {
  version = "~> 1.0"
}

provider "random" {
  version = "~> 1.1"
}

provider "tls" {
  version = "~> 1.0.1"
}

provider "aws" {
  region              = "${var.region}"
  allowed_account_ids = ["${var.allowed_account_ids}"]
  version             = "~> 1.7.1"
}

