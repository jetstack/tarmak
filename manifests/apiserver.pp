# class kubernetes::master
class kubernetes::apiserver(
  $allow_privileged = true,
  $admission_control = undef,
  $count = 1,
  $storage_backend = undef,
  $etcd_nodes = ['localhost'],
  $etcd_port = 2379,
  $etcd_events_port = undef,
  $etcd_ca_file = undef,
  $etcd_cert_file = undef,
  $etcd_key_file = undef,
  $ca_file = undef,
  $cert_file = undef,
  $key_file = undef,
  $systemd_wants = [],
  $systemd_requires = [],
  $systemd_after = [],
  $systemd_before = [],
  $runtime_config = [],
  $insecure_bind_address = undef,
  Array[String] $abac_full_access_users = [],
  Array[String] $abac_read_only_access_users = [],
)  {
  require ::kubernetes

  $_systemd_wants = $systemd_wants
  $_systemd_requires = $systemd_requires
  $_systemd_after = ['network.target'] + $systemd_after
  $_systemd_before = $systemd_before

  # Admission controllers
  if $admission_control == undef {
    # No DefaultStorageClass controller pre 1.4
    if versioncmp($::kubernetes::version, '1.4.0') >= 0 {
      $_admission_control =  ['NamespaceLifecycle', 'LimitRanger', 'ServiceAccount', 'DefaultStorageClass', 'ResourceQuota']
    } else {
      $_admission_control =  ['NamespaceLifecycle', 'LimitRanger', 'ServiceAccount', 'ResourceQuota']

    }
  } else {
    $_admission_control = $admission_control
  }

  # Default to etcd2 for versions bigger than 1.5
  if $storage_backend == undef and versioncmp($::kubernetes::version, '1.5.0') >= 0 {
    $_storage_backend = 'etcd2'
  } else {
    $_storage_backend = $storage_backend
  }

  $service_name = 'kube-apiserver'

  if $etcd_ca_file == undef and $etcd_cert_file == undef and $etcd_key_file == undef {
    $etcd_proto = 'http'
  } else {
    $etcd_proto = 'https'
  }

  $_etcd_urls = map($etcd_nodes) |$node| { "${etcd_proto}://${node}:${etcd_port}" }
  $etcd_servers = $_etcd_urls.join(',')

  if $etcd_events_port == undef {
    $etcd_servers_overrides = []
  }
  else {
    $_etcd_events_urls = map($etcd_nodes) |$node| { "${etcd_proto}://${node}:${etcd_events_port}" }
    $etcd_events_servers = $_etcd_events_urls.join(';')
    $etcd_servers_overrides = [
      "/events#${etcd_events_servers}",
    ]
  }

  $authorization_mode = $kubernetes::_authorization_mode

  # enable alpha RBAC in kubernetes versions before 1.5
  if $runtime_config == [] {
    if member($authorization_mode, 'RBAC') and versioncmp($::kubernetes::version, '1.6.0') < 0 {
      $_runtime_config = [
        'rbac.authorization.k8s.io/v1alpha1=true'
      ]
    } else {
      $_runtime_config = []
    }
  } else {
    $_runtime_config = $runtime_config
  }


  # if ABAC is enabled
  if member($authorization_mode, 'ABAC'){
    # if no full access users are set, set sensible defaults
    if $abac_full_access_users == [] {
      $_abac_full_access_users = [
        'system:serviceaccount:kube-system:default',
        'admin',
        'system:node',
        'system:node:*',
        'system:kube-scheduler',
        'system:kube-controller-manager',
        'system:kube-proxy',
        'system:kube-apiserver'
      ]
    }
    else {
      $_abac_full_access_users = $abac_full_access_users
    }

    # if no read only users are set, set sensible defaults
    if $abac_read_only_access_users == [] and member($authorization_mode, 'ABAC'){
      $_abac_read_only_access_users = ['system:serviceaccount:monitoring:default']
    }
    else {
      $_abac_read_only_access_users = $abac_read_only_access_users
    }

    if versioncmp($::kubernetes::version, '1.5.0') >= 0 {
      $abac_supports_groups = true
    } else {
      $abac_supports_groups = false
    }

    $authorization_policy_file = "${::kubernetes::config_dir}/${service_name}-abac-policy.json"
    file{$authorization_policy_file:
      ensure  => file,
      mode    => '0640',
      owner   => 'root',
      group   => $::kubernetes::params::group,
      content => template("kubernetes/${service_name}-policy.json.erb"),
      require => Kubernetes::Symlink['apiserver'],
      notify  => Service["${service_name}.service"],
    }
  }

  kubernetes::symlink{'apiserver':} ->
  file{"${::kubernetes::systemd_dir}/${service_name}.service":
    ensure  => file,
    mode    => '0644',
    owner   => 'root',
    group   => 'root',
    content => template("kubernetes/${service_name}.service.erb"),
    notify  => Service["${service_name}.service"],
  } ~>
  exec { "${service_name}-daemon-reload":
    command     => 'systemctl daemon-reload',
    path        => $::kubernetes::path,
    refreshonly => true,
  } ->
  service{ "${service_name}.service":
    ensure => running,
    enable => true,
  }

}
