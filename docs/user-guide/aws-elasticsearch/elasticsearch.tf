variable "name" {
  default = "tarmak-logs"
}

variable "region" {
  default = "eu-west-1"
}

provider "aws" {
  region = "${var.region}"
}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "es" {
  statement {
    actions = [
      "es:*",
    ]

    principals {
      type = "AWS"

      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root",
      ]
    }
  }
}

resource "aws_elasticsearch_domain" "es" {
  domain_name           = "${var.name}"
  elasticsearch_version = "6.2"

  cluster_config {
    instance_type = "t2.medium.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_type = "gp2"
    volume_size = 30
  }

  access_policies = "${data.aws_iam_policy_document.es.json}"
}

data "aws_iam_policy_document" "es_shipping" {
  statement {
    actions = [
      "es:ESHttpHead",
      "es:ESHttpPost",
      "es:ESHttpGet",
    ]

    resources = [
      "arn:aws:es:${var.region}:${data.aws_caller_identity.current.account_id}:domain/${var.name}/*",
    ]
  }
}

resource "aws_iam_policy" "es_shipping" {
  name        = "${var.name}-shipping"
  description = "Allows shipping of logs to elasticsearch"

  policy = "${data.aws_iam_policy_document.es_shipping.json}"
}

output "elasticsearch_endpoint" {
  value = "${aws_elasticsearch_domain.es.endpoint}"
}

output "elasticsearch_shipping_policy_arn" {
  value = "${aws_iam_policy.es_shipping.arn}"
}
