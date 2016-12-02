# == Class vault_client::config
#
# This class is called from vault_client for service config.
#
class vault_client::config {

  exec { "In dev mode get CA":
    command => "/bin/bash /etc/sysconfig/vault && /usr/bin/vault vault read -field=certificate $cluster_name/pki/etcd-k8s/cert/ca > /etc/pki/ca-trust/source/anchors/etcd-k8s.pem",
    unless => "/bin/bash /etc/sysconfig/vault && grep ${/usr/bin/vault vault read -field=certificate $cluster_name/pki/etcd-k8s/cert/ca} /etc/pki/ca-trust/source/anchors/etcd-k8s.pem",
    notify => Exec["update CA trust"],
  }

  exec { "update CA trust":
    command => "update-ca-trust",
  }  

}
