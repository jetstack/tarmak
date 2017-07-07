resource "random_id" "consul_encrypt" {
  byte_length = 16
}

resource "random_id" "consul_master_token" {
  byte_length = 32
}
