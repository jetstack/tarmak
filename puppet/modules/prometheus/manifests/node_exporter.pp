class prometheus::node_exporter (
  String $image = 'prom/node-exporter',
  String $version = '0.16.0',
  String $download_url = 'https://github.com/prometheus/node_exporter/releases/download/v#VERSION#/node_exporter-#VERSION#.linux-amd64.tar.gz',
  String $sha256sums_url = 'https://github.com/prometheus/node_exporter/releases/download/v#VERSION#/sha256sums.txt',
  String $signature_url = 'https://releases.tarmak.io/signatures/node_exporter/#VERSION#/sha256sums.txt.asc',
  $port = 9100,
)
{
  include ::prometheus
  $namespace = $::prometheus::namespace

  $ignored_mount_points = '^/(sys|proc|dev|host|etc)($|/)'

  # Setup deployment scrapes and rules for node_exporter
  if $::prometheus::role == 'master' {
    include ::prometheus::server
    $kubernetes_token_file = $::prometheus::server::kubernetes_token_file
    $kubernetes_ca_file = $::prometheus::server::kubernetes_ca_file

    prometheus::rule { 'NodeHighCPUUsage':
      expr        => '(100 - (avg(irate(node_cpu_seconds_total{mode="idle"}[5m])) WITHOUT (cpu) * 100)) > 80',
      for         => '5m',
      summary     => '{{$labels.instance}}: High CPU usage detected',
      description => '{{$labels.instance}}: CPU usage is above 80% (current value is: {{ $value }})',
    }

    prometheus::rule { 'NodeHighLoadAverage':
      expr        => '((node_load5 / count without (cpu, mode) (node_cpu_seconds_total{mode="system"})) > 3)',
      for         => '5m',
      summary     => '{{$labels.instance}}: High load average detected',
      description => '{{$labels.instance}}: 5 minute load average is {{$value}}',
    }

    # TODO: Alert if diskspace is running out in x hours
    prometheus::rule { 'NodeLowDiskSpace':
      expr        => '((node_filesystem_size_bytes - node_filesystem_free_bytes ) / node_filesystem_size_bytes * 100) > 75',
      for         => '2m',
      summary     => '{{$labels.instance}}: Low disk space',
      description => '{{$labels.instance}}: Disk usage is above 75% (current value is: {{ $value }}%)',
    }

    # TODO: Alert when swap is in use
    prometheus::rule { 'NodeSwapEnabled':
      expr        => '(((node_memory_SwapTotal_bytes-node_memory_SwapFree_bytes)/node_memory_SwapTotal_bytes)*100) > 75',
      for         => '2m',
      summary     => '{{$labels.instance}}: Swap usage detected',
      description => '{{$labels.instance}}: Swap usage usage is above 75% (current value is: {{ $value }})',
    }

    prometheus::rule { 'NodeHighMemoryUsage':
      expr        => '(((node_memory_MemTotal_bytes-node_memory_MemFree_bytes-node_memory_Cached_bytes)/(node_memory_MemTotal_bytes)*100)) > 80',
      for         => '5m',
      summary     => '{{$labels.instance}}: High memory usage detected',
      description => '{{$labels.instance}}: Memory usage usage is above 80% (current value is: {{ $value }})',
    }


    # scrape node exporter running on etcd nodes
    prometheus::scrape_config { 'etcd-nodes-exporter':
      order  =>  135,
      config => {
        'dns_sd_configs'  => [{
          'names' => $tarmak::etcd_cluster_exporters,
        }],
        'relabel_configs' => [{
          'source_labels' => ['__address__'],
          'regex'         => '(.+):(.+)',
          'target_label'  => '__address__',
          'replacement'   => '${1}:9100',
        }],
      }
    }

    if $::prometheus::mode == 'Full' {
      $node_ensure = $::prometheus::ensure
    } else {
      $node_ensure = 'absent'
    }

    kubernetes::apply{'node-exporter':
      ensure    => $node_ensure,
      manifests => [
        template('prometheus/prometheus-ns.yaml.erb'),
        template('prometheus/node-exporter-ds.yaml.erb'),
      ],
    }

    # scrape node exporter running on every kubernetes node (through api proxy)
    prometheus::scrape_config { 'kubernetes-nodes-exporter':
      ensure => $node_ensure,
      order  =>  130,
      config => {
        'kubernetes_sd_configs' => [{
          'role' => 'node',
        }],
        'tls_config'            => {
          'ca_file' => $kubernetes_ca_file,
        },
        'bearer_token_file'     => $kubernetes_token_file,
        'scheme'                => 'https',
        'relabel_configs'       => [{
          'action' => 'labelmap',
          'regex'  => '__meta_kubernetes_node_label_(.+)',
          },{
            'target_label' => '__address__',
            'replacement'  => 'kubernetes.default.svc:443',
            }, {
              'source_labels' => ['__meta_kubernetes_node_name'],
              'regex'         => '(.+)',
              'target_label'  => '__metrics_path__',
              'replacement'   => "/api/v1/nodes/\${1}:${port}/proxy/metrics",
          }],
      }
    }
  }

  # Setup node_exporter service on etcd nodes
  if $::prometheus::role == 'etcd' {
    $_download_url = regsubst($download_url, '#VERSION#' ,$version , 'G')
    $_signature_url = regsubst($signature_url, '#VERSION#' ,$version , 'G')
    $_sha256sums_url = regsubst($sha256sums_url, '#VERSION#' ,$version , 'G')

    $dest_dir = "/opt/node_exporter-${version}"
    file { $dest_dir:
      ensure => directory,
      mode   => '0755',
    }
    -> archive { "${dest_dir}/node_exporter.tar.gz":
      ensure            => present,
      extract           => true,
      extract_path      => $dest_dir,
      extract_command   => 'tar xfz %s --strip-components=1',
      source            => $_download_url,
      sha256sums        => $_sha256sums_url,
      signature_armored => $_signature_url,
      provider          =>  'airworthy',
    }
    -> file { "${::prometheus::systemd_path}/node-exporter.service":
      ensure  => $prometheus::ensure,
      content => template('prometheus/node-exporter.service.erb'),
      notify  => Exec["${module_name}-systemctl-daemon-reload"],
    }
    ~> service { 'node-exporter.service':
      ensure  => $::prometheus::service_ensure,
      enable  => $::prometheus::service_enable,
      require => Exec["${module_name}-systemctl-daemon-reload"],
    }
  }
}
