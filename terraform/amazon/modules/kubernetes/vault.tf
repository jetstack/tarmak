resource "tarmak_vault_instance_role" "master" {  
  role_name = "master"
  vault_cluster_name = "${var.vault_cluster_name}"
  internal_fqdns = ["${var.internal_fqdns}"]
  vault_ca = "${var.vault_ca}"
  vault_status = "${var.vault_status}"
}

resource "tarmak_vault_instance_role" "worker" {
  role_name = "worker"
  vault_cluster_name = "${var.vault_cluster_name}"
  internal_fqdns = ["${var.internal_fqdns}"]
  vault_ca = "${var.vault_ca}"
  vault_status = "${var.vault_status}"
}

resource "tarmak_vault_instance_role" "etcd" {
  role_name = "etcd"
  vault_cluster_name = "${var.vault_cluster_name}"
  internal_fqdns = ["${var.internal_fqdns}"]
  vault_ca = "${var.vault_ca}"
  vault_status = "${var.vault_status}"
}