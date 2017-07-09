# CA certificate
resource "tls_private_key" "ca" {
  algorithm = "RSA"
  rsa_bits  = "4096"
}

resource "tls_self_signed_cert" "ca" {
  key_algorithm   = "${tls_private_key.ca.algorithm}"
  private_key_pem = "${tls_private_key.ca.private_key_pem}"

  subject {
    common_name = "Vault ${var.environment} CA"
  }

  is_ca_certificate = true

  # 10 years
  validity_period_hours = 87660

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "cert_signing",
  ]
}

# Per instance certs
resource "tls_private_key" "vault" {
  count = "${var.instance_count}"

  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "tls_cert_request" "vault" {
  count           = "${var.instance_count}"
  key_algorithm   = "${element(tls_private_key.vault.*.algorithm, count.index)}"
  private_key_pem = "${element(tls_private_key.vault.*.private_key_pem, count.index)}"

  subject {
    common_name = "vault-${count.index + 1}.${var.environment}"
  }

  dns_names = [
    "vault.${data.terraform_remote_state.network.private_zone}",
    "vault-${count.index + 1}.${data.terraform_remote_state.network.private_zone}",
    "localhost",
  ]

  ip_addresses = [
    "127.0.0.1",
  ]
}

resource "tls_locally_signed_cert" "vault" {
  count = "${var.instance_count}"

  cert_request_pem = "${element(tls_cert_request.vault.*.cert_request_pem, count.index)}"

  ca_key_algorithm   = "${tls_self_signed_cert.ca.key_algorithm}"
  ca_private_key_pem = "${tls_private_key.ca.private_key_pem}"
  ca_cert_pem        = "${tls_self_signed_cert.ca.cert_pem}"

  # 1 year
  validity_period_hours = 8766

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}
