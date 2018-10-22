resource "tarmak_vault_cluster" "vault" {
  internal_fqdns        = ["${var.internal_fqdns}"]
  vault_ca              = "${var.vault_ca}"
  vault_kms_key_id      = "${var.vault_kms_key_id}"
  vault_unseal_key_name = "${var.vault_unseal_key_name}"
}

resource "tarmak_vault_instance_role" "master" {
  role_name          = "master"
  vault_cluster_name = "${var.vault_cluster_name}"
  internal_fqdns     = ["${var.internal_fqdns}"]
  vault_ca           = "${var.vault_ca}"

  depends_on = ["tarmak_vault_cluster.vault"]
}

resource "tarmak_vault_instance_role" "worker" {
  role_name          = "worker"
  vault_cluster_name = "${var.vault_cluster_name}"
  internal_fqdns     = ["${var.internal_fqdns}"]
  vault_ca           = "${var.vault_ca}"

  depends_on = ["tarmak_vault_cluster.vault"]
}

resource "tarmak_vault_instance_role" "etcd" {
  role_name          = "etcd"
  vault_cluster_name = "${var.vault_cluster_name}"
  internal_fqdns     = ["${var.internal_fqdns}"]
  vault_ca           = "${var.vault_ca}"

  depends_on = ["tarmak_vault_cluster.vault"]
}
