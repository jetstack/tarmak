##exporter_# Sets up a blackbox exporter to forward etcd metrics from etcd nodes
class prometheus::blackbox_exporter_etcd (
  String $download_url = 'https://github.com/jetstack-experimental/blackbox_exporter/releases/download/v#VERSION#/blackbox_exporter_#VERSION#_linux_amd64',
  String $version = '0.4.0-jetstack',
  String $config_dir = '/etc/blackbox_exporter',
  Integer $port = 9115,
)
{
  include ::prometheus

  $dest_dir = "/opt/blackbox_exporter-${version}"
  $systemd_path = $::prometheus::systemd_path

  $etcd_k8s_main_port = $::prometheus::etcd_k8s_main_port
  $etcd_k8s_events_port = $::prometheus::etcd_k8s_events_port
  $etcd_overlay_port = $::prometheus::etcd_overlay_port

  # Setup scrapes and rules for blackbox_etcd
  if $::prometheus::role == 'master' {
    include ::prometheus::server

    prometheus::scrape_config { 'etcd-k8s-main':
      order  =>  140,
      config => {
        'metrics_path'    => '/probe',
        'params'          => { 'module' => ['etcd_k8s_main_proxy'] },
        'dns_sd_configs'  => [{
          'names' => $tarmak::etcd_cluster_exporters,
        }],
        'relabel_configs' => [{
          'source_labels' => [],
          'regex'         => '(.*)',
          'target_label'  => '__param_target',
          'replacement'   => "https://127.0.0.1:${etcd_k8s_main_port}/metrics",
        }],
      }
    }

    prometheus::scrape_config { 'etcd-k8s-events':
      order  =>  150,
      config => {
        'metrics_path'    => '/probe',
        'params'          => { 'module' => ['etcd_k8s_events_proxy'] },
        'dns_sd_configs'  => [{
          'names' => $tarmak::etcd_cluster_exporters,
        }],
        'relabel_configs' => [{
          'source_labels' => [],
          'regex'         => '(.*)',
          'target_label'  => '__param_target',
          'replacement'   => "https://127.0.0.1:${etcd_k8s_events_port}/metrics",
        }],
      }
    }

    prometheus::scrape_config { 'etcd-overlay':
      order  =>  160,
      config => {
        'metrics_path'    => '/probe',
        'params'          => { 'module' => ['etcd_overlay_proxy'] },
        'dns_sd_configs'  => [{
          'names' => $tarmak::etcd_cluster_exporters,
        }],
        'relabel_configs' => [{
          'source_labels' => [],
          'regex'         => '(.*)',
          'target_label'  => '__param_target',
          'replacement'   => "https://127.0.0.1:${etcd_overlay_port}/metrics",
        }],
      }
    }

    prometheus::rule { 'EtcdDown':
      # TODO: we should limit this on the etcd jobs
      expr        => '(probe_success !=1 AND probe_success{instance=~".*etcd.*"})',
      for         => '2m',
      summary     => '{{$labels.instance}}: etcd server probe failed',
      description => '{{$labels.instance}}: etcd server probe failed for {{$labels.job}}',
      order       => 10,
    }

    prometheus::rule { 'EtcdNoLeader':
      # TODO: we should limit this on the etcd jobs
      expr        => '(etcd_server_has_leader != 1)',
      for         => '2m',
      summary     => '{{$labels.instance}}: etcd server has no leader',
      description => '{{$labels.instance}}: etcd cluster server has no leader',
      order       => 10,
    }
  }

  # Setup blackbox service on etcd nodes
  if $::prometheus::role == 'etcd' {
    $_download_url = regsubst($download_url, '#VERSION#' ,$version , 'G')

    file { $dest_dir:
      ensure => directory,
      mode   => '0755',
    }
    -> archive { "${dest_dir}/blackbox_exporter":
      ensure   => present,
      extract  => false,
      source   => $_download_url,
      provider => 'airworthy',
    }

    file { $config_dir:
      ensure => directory,
      mode   => '0755',
    }
    -> file { "${config_dir}/blackbox_exporter.yaml":
      ensure  => file,
      content => template('prometheus/blackbox_exporter.yaml.erb'),
    }

    file { "${systemd_path}/blackbox-exporter.service":
      ensure  => file,
      content => template('prometheus/blackbox_exporter.service.erb'),
    }
    ~> exec { "${module_name}-systemctl-daemon-reload":
      command     => '/usr/bin/systemctl daemon-reload',
      refreshonly => true,
    }

    service { 'blackbox-exporter':
      ensure    => running,
      enable    => true,
      subscribe => [
        Archive["${dest_dir}/blackbox_exporter"],
        File["${config_dir}/blackbox_exporter.yaml"],
        File["${systemd_path}/blackbox-exporter.service"]
      ],
    }
  }
}
