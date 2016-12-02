# == Class vault_client::config
#
# This class is called from vault_client for service config.
#
class vault_client::config {

  exec { "In dev mode get CA":
    command => "/bin/bash -c 'source /etc/sysconfig/vault; /usr/bin/vault read -address=\$VAULT_ADDR -field=certificate \$CLUSTER_NAME/pki/etcd-k8s/cert/ca > /etc/pki/ca-trust/source/anchors/etcd-k8s.pem'",
    onlyif  => "/bin/bash -c 'source /etc/sysconfig/vault; /usr/bin/vault read -address=\$VAULT_ADDR -field=certificate \$CLUSTER_NAME/pki/etcd-k8s/cert/ca | diff - /etc/pki/ca-trust/source/anchors/etcd-k8s.pem'",
    notify => Exec["update CA trust"],
  }

  exec { "update CA trust":
    command => "/usr/bin/update-ca-trust",
    refreshonly => true,
  }  

}
