# == Class vault_client::config
#
# This class is called from vault_client for service config.
#
class vault_client::config {

  file { '/etc/sysconfig/vault':
    ensure  => file,
    content => template('vault_client/vault.erb'),
  }

  file { [ '/etc/etcd', '/etc/etcd/ssl', '/etc/etcd/ssl/certs' ]:
    ensure => directory,
  }

  user { 'etcd user for vault':
    ensure => present,
    name   => 'etcd',
    uid    => 873,
    shell  => '/sbin/nologin',
    home   => '/var/lib/etcd',
  }

  if $vault_client::role == 'master' or $vault_client::role == 'worker' {

  file { [ '/etc/kubernetes', '/etc/kubernetes/ssl', '/etc/kubernetes/ssl/certs' ]:
    ensure => directory,
  }

  user { 'k8s user for vault':
      ensure => present,
      name   => 'k8s',
      uid    => 837,
      shell  => '/sbin/nologin',
      home   => '/var/lib/kubernetes',
    }
  }

  if $vault_client::role == 'master' or $vault_client::role == 'etcd' {
    exec { 'In dev mode get CA for k8s':
      command => "/bin/bash -c 'source /etc/sysconfig/vault; /usr/bin/vault read -address=\$VAULT_ADDR -field=certificate \$CLUSTER_NAME/pki/etcd-k8s/cert/ca > /etc/etcd/ssl/certs/etcd-k8s.pem'",
      unless  => "/bin/bash -c 'source /etc/sysconfig/vault; /usr/bin/vault read -address=\$VAULT_ADDR -field=certificate \$CLUSTER_NAME/pki/etcd-k8s/cert/ca | diff -P /etc/etcd/ssl/certs/etcd-k8s.pem -'",
      require => File['/etc/etcd/ssl/certs'],
    }

    vault_client::etcd_cert_service { 'k8s':
      etcd_cluster => 'k8s',
      frequency    => '1d',
      notify       => Exec['Trigger k8s cert'],
      require      => [ File['/etc/etcd/ssl'], User['etcd user for vault'] ],
    }

    service { 'etcd-k8s-cert.timer':
      provider => systemd,
      enable   => true,
      require  => [ File['/usr/lib/systemd/system/etcd-k8s-cert.timer'], Exec['In dev mode get CA for k8s'] ],
    }
  }

  exec { 'Trigger k8s cert':
    command     => '/usr/bin/systemctl start etcd-k8s-cert.service',
    user        => 'root',
    unless      => '/usr/bin/openssl x509 -checkend 3600 -in /etc/etcd/ssl/certs/etcd-k8s-cert.pem | /usr/bin/grep "Certificate will not expire"',
    #require     => File['/usr/lib/systemd/system/etcd-k8s-cert.service'],
  }

  exec { 'In dev mode get CA for overlay':
    command => "/bin/bash -c 'source /etc/sysconfig/vault; /usr/bin/vault read -address=\$VAULT_ADDR -field=certificate \$CLUSTER_NAME/pki/etcd-overlay/cert/ca > /etc/etcd/ssl/certs/etcd-overlay.pem'",
    unless  => "/bin/bash -c 'source /etc/sysconfig/vault; /usr/bin/vault read -address=\$VAULT_ADDR -field=certificate \$CLUSTER_NAME/pki/etcd-overlay/cert/ca | diff -P /etc/etcd/ssl/certs/etcd-overlay.pem -'",
    require => File['/etc/etcd/ssl/certs'],
  }

  #not used for now
  exec { 'update CA trust':
    command     => '/usr/bin/update-ca-trust',
    refreshonly => true,
  }

  vault_client::etcd_cert_service { 'overlay':
    etcd_cluster => 'overlay',
    frequency    => '1d',
    notify       => Exec['Trigger overlay cert'],
    require      => [ File['/etc/etcd/ssl'], User['etcd user for vault'] ],
  }

  service { 'etcd-overlay-cert.timer':
    provider => systemd,
    enable   => true,
    require  => [ File['/usr/lib/systemd/system/etcd-overlay-cert.timer'], Exec['In dev mode get CA for overlay'] ],
  }

  exec { 'Trigger overlay cert':
    command     => '/usr/bin/systemctl start etcd-overlay-cert.service',
    user        => 'root',
    unless      => '/usr/bin/stat /etc/etcd/ssl/certs/etcd-overlay-cert.pem || /usr/bin/openssl x509 -checkend 3600 -in /etc/etcd/ssl/certs/etcd-overlay-cert.pem | /usr/bin/grep "Certificate will not expire"',
    require     => File['/usr/lib/systemd/system/etcd-overlay-cert.service'],
  }

  exec { 'Trigger events cert':
    command     => '/usr/bin/systemctl start etcd-events-cert.service',
    user        => 'root',
    unless      => '/usr/bin/openssl x509 -checkend 3600 -in /etc/etcd/ssl/certs/etcd-events-cert.pem | /usr/bin/grep "Certificate will not expire"',
    #require     => File['/usr/lib/systemd/system/etcd-events-cert.service'],
  }


  if $vault_client::role == 'worker' {
    vault_client::k8s_cert_service { 'kubelet':
      k8s_component => 'kubelet',
      frequency     => '1d',
      notify        => Exec['Trigger kubelet cert'],
      require       => [ File['/etc/kubernetes/ssl'], User['k8s user for vault'] ],
    }

    exec { 'Trigger kubelet cert':
      command     => '/usr/bin/systemctl start k8s-kubelet-cert.service',
      user        => 'root',
      refreshonly => true,
    }
  }
}
