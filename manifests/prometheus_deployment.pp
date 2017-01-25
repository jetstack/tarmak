class prometheus::prometheus_deployment (
  $addon_dir = $::prometheus::addon_dir,
  $helper_dir = $::prometheus::helper_dir,
  $prometheus_namespace = $::prometheus::prometheus_namespace,
  $prometheus_image = $::prometheus::prometheus_image,
  $prometheus_version = $::prometheus::prometheus_version,
  $prometheus_storage_local_retention = $::prometheus::prometheus_storage_local_retention,
  $prometheus_storage_local_memchunks = $::prometheus::prometheus_storage_local_memchunks,
  $prometheus_port = $::prometheus::prometheus_port,
  $prometheus_use_module_config = $::prometheus::prometheus_use_module_config,
  $etcd_cluster = $::prometheus::etcd_cluster,
  $etcd_k8s_port = $::prometheus::etcd_k8s_port,
  $etcd_events_port = $::prometheus::etcd_events_port,
  $etcd_overlay_port = $::prometheus::etcd_overlay_port,
  $prometheus_use_module_rules = $::prometheus::prometheus_use_module_rules,
  $prometheus_install_state_metrics = $::prometheus::prometheus_install_state_metrics,
  $prometheus_install_node_exporter = $::prometheus::prometheus_install_node_exporter
)
{
  file { "${helper_dir}/kubectl_helper_prom.sh":
    ensure  => file,
    content => template('prometheus/kubectl_helper.sh.erb'),
    mode    => '0755',
  }

  if $prometheus_use_module_config {
    file { "${addon_dir}/prometheus-config.yaml":
      ensure  => file,
      content => template('prometheus/prometheus-config.yaml.erb'),
    } ->
    exec { 'Install prom-config':
      command => "${helper_dir}/kubectl_helper_prom.sh apply ${addon_dir}/prometheus-config.yaml",
      unless  => "${helper_dir}/kubectl_helper_prom.sh get ${addon_dir}/prometheus-config.yaml",
      before  => Exec['Install prom-deploy'],
      require => File["${helper_dir}/kubectl_helper_prom.sh"],
    }
  }

  if $prometheus_use_module_rules {
    file { "${addon_dir}/prometheus-rules.yaml":
      ensure  => file,
      content => template('prometheus/prometheus-rules.yaml.erb'),
    } ->
    exec { 'Install prom-rules':
      command => "${helper_dir}/kubectl_helper_prom.sh apply ${addon_dir}/prometheus-rules.yaml",
      unless  => "${helper_dir}/kubectl_helper_prom.sh get ${addon_dir}/prometheus-rules.yaml",
      before  => Exec['Install prom-deploy'],
      require => File["${helper_dir}/kubectl_helper_prom.sh"],
    }
  }

  file { "${addon_dir}/prometheus-deployment.yaml":
    ensure  => file,
    content => template('prometheus/prometheus-deployment.yaml.erb'),
  } ->
  exec { 'Install prom-deploy':
    command => "${helper_dir}/kubectl_helper_prom.sh apply ${addon_dir}/prometheus-deployment.yaml",
    unless  => "${helper_dir}/kubectl_helper_prom.sh get ${addon_dir}/prometheus-deployment.yaml",
    require => File["${helper_dir}/kubectl_helper_prom.sh"],
  }

  file { "${addon_dir}/prometheus-svc.yaml":
    ensure  => file,
    content => template('prometheus/prometheus-svc.yaml.erb'),
  } ->
  exec { 'Install prom-svc':
    command => "${helper_dir}/kubectl_helper_prom.sh apply ${addon_dir}/prometheus-svc.yaml",
    unless  => "${helper_dir}/kubectl_helper_prom.sh get ${addon_dir}/prometheus-svc.yaml",
    require => File["${helper_dir}/kubectl_helper_prom.sh"],
  }

  if $prometheus_install_state_metrics {
    file { "${addon_dir}/kube-state-metrics-deployment.yaml":
      ensure  => file,
      content => template('prometheus/kube-state-metrics-deployment.yaml.erb'),
    } ->
    exec { 'Install kube-state-metrics':
      command => "${helper_dir}/kubectl_helper_prom.sh apply ${addon_dir}/kube-state-metrics-deployment.yaml",
      unless  => "${helper_dir}/kubectl_helper_prom.sh get ${addon_dir}/kube-state-metrics-deployment.yaml",
      require => File["${helper_dir}/kubectl_helper_prom.sh"],
    }
  }

  if $prometheus_install_node_exporter {
    file { "${addon_dir}/prometheus-node-exporter-ds.yaml":
      ensure  => file,
      content => template('prometheus/prometheus-node-exporter-ds.yaml.erb'),
    } ->
    exec { 'Install prom-node-exporter':
      command => "${helper_dir}/kubectl_helper_prom.sh apply ${addon_dir}/prometheus-node-exporter-ds.yaml",
      unless  => "${helper_dir}/kubectl_helper_prom.sh get ${addon_dir}/prometheus-node-exporter-ds.yaml",
      require => File["${helper_dir}/kubectl_helper_prom.sh"],
    }
  }
}
