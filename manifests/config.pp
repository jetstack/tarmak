# == Class vault_client::config
#
# This class is called from vault_client for service config.
#
class vault_client::config {

  file { "/etc/sysconfig/vault":
    ensure   => file,
    content  => template('vault_client/vault.erb'),
  }

  file { "etcd working dir for vault":
    path => "/etc/etcd",
    ensure => directory,
  }

  user { "etcd user for vault":
    name => "etcd",
    ensure => present,
    uid => 873,
    shell => "/sbin/nologin",
    home => "/var/lib/etcd",
  }

  exec { "In dev mode get CA for k8s":
    command => "/bin/bash -c 'source /etc/sysconfig/vault; /usr/bin/vault read -address=\$VAULT_ADDR -field=certificate \$CLUSTER_NAME/pki/etcd-k8s/cert/ca > /etc/pki/ca-trust/source/anchors/etcd-k8s.pem'",
    unless  => "/bin/bash -c 'source /etc/sysconfig/vault; /usr/bin/vault read -address=\$VAULT_ADDR -field=certificate \$CLUSTER_NAME/pki/etcd-k8s/cert/ca | diff -P /etc/pki/ca-trust/source/anchors/etcd-k8s.pem -'",
    notify  => Exec["update CA trust"],
  }

  exec { "In dev mode get CA for overlay":
    command => "/bin/bash -c 'source /etc/sysconfig/vault; /usr/bin/vault read -address=\$VAULT_ADDR -field=certificate \$CLUSTER_NAME/pki/etcd-overlay/cert/ca > /etc/etcd/ssl/certs/etcd-overlay.pem'",
    unless  => "/bin/bash -c 'source /etc/sysconfig/vault; /usr/bin/vault read -address=\$VAULT_ADDR -field=certificate \$CLUSTER_NAME/pki/etcd-overlay/cert/ca | diff -P /etc/etcd/ssl/certs/etcd-overlay.pem -'",
  }

  exec { "update CA trust":
    command => "/usr/bin/update-ca-trust",
    refreshonly => true,
  }  

  vault_client::etcd_cert_service { "k8s":
    etcd_cluster => "k8s",
    frequency    => "1d",
    notify       => Exec["Trigger k8s cert"],
    require      => [ File["etcd working dir for vault"], User["etcd user for vault"] ],
  }

  service { "etcd-k8s-cert.timer":
    provider => systemd,
    enable   => true,
    require  => [ File['/usr/lib/systemd/system/etcd-k8s-cert.timer'], Exec['In dev mode get CA for k8s'] ],
  }

  exec { "Trigger k8s cert":
    command => "/usr/bin/systemctl start etcd-k8s-cert.service",
    user => "root",
    refreshonly => true,
  }

  vault_client::etcd_cert_service { "overlay":
    etcd_cluster => "overlay",
    frequency    => "1d",
    notify       => Exec["Trigger overlay cert"],
    require      => [ File["etcd working dir for vault"], User["etcd user for vault"] ],
  }

  service { "etcd-overlay-cert.timer":
    provider => systemd,
    enable   => true,
    require  => [ File['/usr/lib/systemd/system/etcd-overlay-cert.timer'], Exec['In dev mode get CA for overlay'] ],
  }


  exec { "Trigger overlay cert":
    command => "/usr/bin/systemctl start etcd-overlay-cert.service",
    user => "root",
    refreshonly => true,
  }

  exec { "Trigger events cert":
    command => "/usr/bin/systemctl start etcd-events-cert.service",
    user => "root",
    refreshonly => true,
  }
}
