provider "aws" {
  region              = "${var.region}"
  allowed_account_ids = ["${var.allowed_account_ids}"]
  version             = "~> 1.0"
}

provider "template" {
  version = "~> 1.0"
}
