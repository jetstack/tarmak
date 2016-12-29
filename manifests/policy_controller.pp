class calico::policy_controller {
  file { '/root/calico-config.yaml':
    ensure  => file,
    content => template('calico/calico-config.yaml.erb'),
  }

  file { '/root/policy-controller-deployment.yaml':
    ensure  => file,
    content => template('calico/policy-controller-deployment.yaml.erb'),
  }

  exec { 'deploy calico config':
    command => '/usr/bin/kubectl apply -f /root/calico-config.yaml',
    unless  => '/usr/bin/kubectl get -f /root/calico-config.yaml',
    require => File['/root/calico-config.yaml'],
  }

  exec { 'deploy calico policy controller':
    command => '/usr/bin/kubectl apply -f /root/policy-controller-deployment.yaml',
    unless  => '/usr/bin/kubectl get -f /root/policy-controller-deployment.yaml',
    require => [ Exec['deploy calico config'], File['/root/policy-controller-deployment.yaml'] ],
  }
}
