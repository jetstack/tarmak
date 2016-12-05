# == Class vault_client::config
#
# This class is called from vault_client for service config.
#
class vault_client::config {

  file { "/etc/sysconfig/vault":
    ensure   => file,
    content  => template('vault_client/vault.erb'),
  }

  exec { "In dev mode get CA":
    command => "/bin/bash -c 'source /etc/sysconfig/vault; /usr/bin/vault read -address=\$VAULT_ADDR -field=certificate \$CLUSTER_NAME/pki/etcd-k8s/cert/ca > /etc/pki/ca-trust/source/anchors/etcd-k8s.pem'",
    unless  => "/bin/bash -c 'source /etc/sysconfig/vault; /usr/bin/vault read -address=\$VAULT_ADDR -field=certificate \$CLUSTER_NAME/pki/etcd-k8s/cert/ca | diff - /etc/pki/ca-trust/source/anchors/etcd-k8s.pem'",
    notify  => Exec["update CA trust"],
  }

  exec { "update CA trust":
    command => "/usr/bin/update-ca-trust",
    refreshonly => true,
  }  

  vault_client::etcd_cert_service { "k8s":
    etcd_cluster => "k8s",
    frequency    => "1d",
    notify       => Exec["Trigger k8s cert"],
  }

  service { "etcd-k8s-cert.timer":
    provider => systemd,
    enable   => true,
    require  => [ File['/usr/lib/systemd/system/etcd-k8s-cert.timer'], Exec['In dev mode get CA'] ],
    notify   => Exec["Trigger k8s cert"],
  }

  exec { "Trigger k8s cert":
    command => "/usr/bin/systemctl start etcd-k8s-cert.service",
    user => "root",
    refreshonly => true,
  }

  vault_client::etcd_cert_service { "overlay":
    etcd_cluster => "overlay",
    frequency    => "1d",
  }
}
