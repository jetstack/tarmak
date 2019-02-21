class prometheus::server (
  String $image = 'prom/prometheus',
  String $version = '2.3.2',
  String $reloader_image = 'jimmidyson/configmap-reload',
  String $reloader_version = '0.1',
  String $retention = '720h',  # 30 days
  Integer[1025,65535] $port = 9090,
  String $external_url = '',
  Boolean $persistent_volume = false,
  Integer $persistent_volume_size = 15,
  String $kubernetes_token_file = '/var/run/secrets/kubernetes.io/serviceaccount/token',
  String $kubernetes_ca_file = '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
  Hash[String,String] $external_labels = {},
)
{
  require ::kubernetes
  include ::prometheus

  if $::prometheus::mode == 'Full' or $::prometheus::mode == 'ExternalScrapeTargetsOnly' {
    $namespace = $::prometheus::namespace

    $authorization_mode = $::kubernetes::_authorization_mode
    if member($authorization_mode, 'RBAC'){
      $rbac_enabled = true
    } else {
      $rbac_enabled = false
    }

    if versioncmp($::kubernetes::version, '1.6.0') >= 0 {
      $version_before_1_6 = false
    } else {
      $version_before_1_6 = true
    }

    kubernetes::apply{'prometheus-server':
      ensure    => $::prometheus::ensure,
      manifests => [
        template('prometheus/prometheus-ns.yaml.erb'),
        template('prometheus/prometheus-deployment.yaml.erb'),
        template('prometheus/prometheus-svc.yaml.erb'),
      ],
    }

    kubernetes::apply{'prometheus-config':
      ensure => $::prometheus::ensure,
      type   => 'concat',
    }

    kubernetes::apply_fragment { 'prometheus-config-header':
      ensure  => $::prometheus::ensure,
      content => template('prometheus/prometheus-config-header.yaml.erb'),
      order   => 0,
      target  => 'prometheus-config',
    }

    kubernetes::apply_fragment { 'prometheus-config-prometheus-file':
      ensure  => $::prometheus::ensure,
      content => '  prometheus.yaml: |-',
      order   => 100,
      target  => 'prometheus-config',
    }

    kubernetes::apply_fragment { 'prometheus-config-prometheus-rules':
      ensure  => $::prometheus::ensure,
      content => template('prometheus/prometheus-config-rules.yaml.erb'),
      order   => 200,
      target  => 'prometheus-config',
    }

    kubernetes::apply_fragment { 'prometheus-config-global':
      ensure  => $::prometheus::ensure,
      content => template('prometheus/prometheus-config-global.yaml.erb'),
      order   => 300,
      target  => 'prometheus-config',
    }

    kubernetes::apply_fragment { 'prometheus-config-global-pre-scrape-config':
      ensure  => $::prometheus::ensure,
      content => '    scrape_configs:',
      order   => 400,
      target  => 'prometheus-config',
    }

    if $::prometheus::mode == 'Full' {
      # Scrape config for API servers.
      #
      # Kubernetes exposes API servers as endpoints to the default/kubernetes
      # service so this uses `endpoints` role and uses relabelling to only keep
      # the endpoints associated with the default/kubernetes service using the
      # default named port `https`. This works for single API server deployments as
      # well as HA API server deployments.
      prometheus::scrape_config { 'kubernetes-apiservers':
        order  =>  100,
        config => {
          'kubernetes_sd_configs' => [{
            'role' => 'endpoints',
          }],
          'tls_config'            => {
            'ca_file' => $kubernetes_ca_file,
          },
          'bearer_token_file'     => $kubernetes_token_file,
          'scheme'                => 'https',
          'relabel_configs'       => [{
            'source_labels' => ['__meta_kubernetes_namespace', '__meta_kubernetes_service_name', '__meta_kubernetes_endpoint_port_name'],
            'action'        => 'keep',
            'regex'         => 'default;kubernetes;https',
          }],
        }
      }

      # Scrape config for master's schedulers and controller manager (kubelet).
      #
      # Rather than connecting directly to the node, the scrape is proxied though the
      # Kubernetes apiserver.  This means it will work if Prometheus is running out of
      # cluster, or can't connect to nodes for some other reason (e.g. because of
      # firewalling).
      prometheus::scrape_config { 'kubernetes-schedulers':
        order  =>  110,
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
              'source_labels' => ['__meta_kubernetes_node_label_role'],
              'action'        => 'keep',
              'regex'         => 'master',
          },{
            'action' => 'labelmap',
            'regex'  => '__meta_kubernetes_node_label_(.+)',
          },{
            'target_label' => '__address__',
            'replacement'  => 'kubernetes.default.svc:443',
          }, {
            'source_labels' => ['__meta_kubernetes_node_name'],
            'regex'         => '(.+)',
            'target_label'  => '__metrics_path__',
            'replacement'   => '/api/v1/nodes/${1}:10251/proxy/metrics',
          }],
        }
      }
      prometheus::scrape_config { 'kubernetes-controller-managers':
        order  =>  110,
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
            'source_labels' => ['__meta_kubernetes_node_label_role'],
            'action'        => 'keep',
            'regex'         => 'master',
          },{
            'action' => 'labelmap',
            'regex'  => '__meta_kubernetes_node_label_(.+)',
          },{
            'target_label' => '__address__',
            'replacement'  => 'kubernetes.default.svc:443',
          }, {
            'source_labels' => ['__meta_kubernetes_node_name'],
            'regex'         => '(.+)',
            'target_label'  => '__metrics_path__',
            'replacement'   => '/api/v1/nodes/${1}:10252/proxy/metrics',
          }],
        }
      }

      # Scrape config for nodes (kubelet).
      #
      # Rather than connecting directly to the node, the scrape is proxied though the
      # Kubernetes apiserver.  This means it will work if Prometheus is running out of
      # cluster, or can't connect to nodes for some other reason (e.g. because of
      # firewalling).
      prometheus::scrape_config { 'kubernetes-nodes':
        order  =>  110,
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
            'replacement'   => '/api/v1/nodes/${1}/proxy/metrics',
          }],
        }
      }


      # Scrape config for Kubelet cAdvisor.
      #
      # This is required for Kubernetes 1.7.3 and later, where cAdvisor metrics
      # (those whose names begin with 'container_') have been removed from the
      # Kubelet metrics endpoint.  This job scrapes the cAdvisor endpoint to
      # retrieve those metrics.
      #
      # In Kubernetes 1.7.0-1.7.2, these metrics are only exposed on the cAdvisor
      # HTTP endpoint; use "replacement: /api/v1/nodes/${1}:4194/proxy/metrics"
      # in that case (and ensure cAdvisor's HTTP server hasn't been disabled with
      # the --cadvisor-port=0 Kubelet flag).
      #
      # This job is not necessary and should be removed in Kubernetes 1.6 and
      # earlier versions, or it will cause the metrics to be scraped twice.
      prometheus::scrape_config { 'kubernetes-nodes-cadvisor':
        order  =>  120,
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
            'replacement'   => '/api/v1/nodes/${1}/proxy/metrics/cadvisor',
          }],
        }
      }

      # Scrape config for service endpoints.
      #
      # The relabeling allows the actual service scrape endpoint to be configured
      # via the following annotations:
      #
      # * `prometheus.io/scrape`: Only scrape services that have a value of `true`
      # * `prometheus.io/scheme`: If the metrics endpoint is secured then you will need
      # to set this to `https` & most likely set the `tls_config` of the scrape config.
      # * `prometheus.io/path`: If the metrics path is not `/metrics` override this.
      # * `prometheus.io/port`: If the metrics are exposed on a different port to the
      # service then set this appropriately.
      prometheus::scrape_config { 'kubernetes-service-endpoints':
        order  =>  200,
        config => {
          'kubernetes_sd_configs' => [{
            'role' => 'endpoints',
          }],

          'relabel_configs'       => [
            {
              'source_labels' => ['__meta_kubernetes_service_annotation_prometheus_io_scrape'],
              'action'        => 'keep',
              'regex'         => true,
            },
            {
              'source_labels' => ['__meta_kubernetes_service_annotation_prometheus_io_scheme'],
              'action'        => 'replace',
              'target_label'  => '__scheme__',
              'regex'         => '(https?)',
            },
            {
              'source_labels' => ['__meta_kubernetes_service_annotation_prometheus_io_path'],
              'action'        => 'replace',
              'target_label'  => '__metrics_path__',
              'regex'         => '(.+)',
            },
            {
              'source_labels' => ['__address__', '__meta_kubernetes_service_annotation_prometheus_io_port'],
              'action'        => 'replace',
              'target_label'  => '__address__',
              'regex'         => '(.+)(?::\d+);(\d+)',
              'replacement'   => '$1:$2',
            },
            {
              'action' => 'labelmap',
              'regex'  => '__meta_kubernetes_service_label_(.+)',
            },
            {
              'source_labels' => ['__meta_kubernetes_namespace'],
              'action'        => 'replace',
              'target_label'  => 'kubernetes_namespace',
            },
            {
              'source_labels' => ['__meta_kubernetes_service_name'],
              'action'        => 'replace',
              'target_label'  => 'kubernetes_name',
            }
          ]
        }
      }

      # Example scrape config for probing services via the Blackbox Exporter.
      #
      # The relabeling allows the actual service scrape endpoint to be configured
      # via the following annotations:
      #
      # * `prometheus.io/probe`: Only probe services that have a value of `true`
      prometheus::scrape_config { 'kubernetes-services':
        order  =>  210,
        config => {
          'kubernetes_sd_configs' => [{
            'role' => 'service'
          }],
          'metrics_path'          => '/probe',
          'params'                => {
            'module' => ['http_2xx'],
          },
          'relabel_configs'       => [
            {
              'source_labels' => ['__meta_kubernetes_service_annotation_prometheus_io_probe'],
              'action'        => 'keep',
              'regex'         => true,
            },
            {
              'source_labels' => ['__address__'],
              'target_label'  => '__param_target',
            },
            {
              'target_label' => '__address__',
              'replacement'  => 'blackbox-exporter',
            },
            {
              'source_labels' => ['__param_target'],
              'target_label'  => 'instance',
            },
            {
              'action' => 'labelmap',
              'regex'  => '__meta_kubernetes_service_label_(.+)',
            },
            {
              'source_labels' => ['__meta_kubernetes_service_namespace'],
              'target_label'  => 'kubernetes_namespace',
            },
            {
              'source_labels' => ['__meta_kubernetes_service_name'],
              'target_label'  => 'kubernetes_name',
            },
          ]
        },
      }


      # Example scrape config for probing ingresses via the Blackbox Exporter.
      #
      # The relabeling allows the actual ingress scrape endpoint to be configured
      # via the following annotations:
      #
      # * `prometheus.io/probe`: Only probe ingresses that have a value of `true`
      prometheus::scrape_config { 'kubernetes-ingresses':
        order  =>  220,
        config => {
          'kubernetes_sd_configs' => [{
            'role' => 'ingress'
          }],
          'metrics_path'          => '/probe',
          'params'                => {
            'module' => ['http_2xx'],
          },
          'relabel_configs'       => [
            {
              'source_labels' => ['__meta_kubernetes_ingress_annotation_prometheus_io_probe'],
              'action'        => 'keep',
              'regex'         => true,
            },
            {
              'source_labels' => ['__meta_kubernetes_ingress_scheme','__address__','__meta_kubernetes_ingress_path'],
              'target_label'  => '__param_target',
              'regex'         => '(.+);(.+);(.+)',
              'replacement'   => '${1}://${2}${3}',
            },
            {
              'target_label' => '__address__',
              'replacement'  => 'blackbox-exporter',
            },
            {
              'source_labels' => ['__param_target'],
              'target_label'  => 'instance',
            },
            {
              'action' => 'labelmap',
              'regex'  => '__meta_kubernetes_ingress_label_(.+)',
            },
            {
              'source_labels' => ['__meta_kubernetes_ingress_namespace'],
              'target_label'  => 'kubernetes_namespace',
            },
            {
              'source_labels' => ['__meta_kubernetes_ingress_name'],
              'target_label'  => 'kubernetes_name',
            },
          ]
        },
      }

      # Example scrape config for pods
      #
      # The relabeling allows the actual pod scrape endpoint to be configured via the
      # following annotations:
      #
      # * `prometheus.io/scrape`: Only scrape pods that have a value of `true`
      # * `prometheus.io/path`: If the metrics path is not `/metrics` override this.
      # * `prometheus.io/port`: Scrape the pod on the indicated port instead of the
      # pod's declared ports (default is a port-free target if none are declared).
      prometheus::scrape_config { 'kubernetes-pods':
        order  =>  230,
        config => {
          'kubernetes_sd_configs' => [{
            'role' => 'pod'
          }],
          'relabel_configs'       => [
            {
              'source_labels' => ['__meta_kubernetes_pod_annotation_prometheus_io_scrape'],
              'action'        => 'keep',
              'regex'         => true,
            },
            {
              'source_labels' => ['__meta_kubernetes_pod_annotation_prometheus_io_path'],
              'action'        => 'replace',
              'target_label'  => '__metrics_path__',
              'regex'         => '(.+)',
            },
            {
              'source_labels' => ['__address__', '__meta_kubernetes_pod_annotation_prometheus_io_port'],
              'action'        => 'replace',
              'regex'         => '([^:]+)(?::\d+)?;(\d+)',
              'replacement'   => '${1}:${2}',
              'target_label'  => '__address__',
            },
            {
              'action' => 'labelmap',
              'regex'  => '__meta_kubernetes_pod_label_(.+)',
            },
            {
              'source_labels' => ['__meta_kubernetes_namespace'],
              'action'        => 'replace',
              'target_label'  => 'kubernetes_namespace',
            },
            {
              'source_labels' => ['__meta_kubernetes_pod_name'],
              'action'        => 'replace',
              'target_label'  => 'kubernetes_pod_name',
            }
          ]
        }
      }
    }

    kubernetes::apply{'prometheus-rules':
      ensure => $::prometheus::ensure,
      type   => 'concat',
    }

    kubernetes::apply_fragment { 'prometheus-rules-header':
      ensure  => $::prometheus::ensure,
      content => template('prometheus/prometheus-rules-header.yaml.erb'),
      order   => 0,
      target  => 'prometheus-rules',
    }


    prometheus::rule { 'ScrapeEndpointDown':
      expr        => '(up == 0 AND up {job != "kubernetes-apiservers"})',
      for         => '2m',
      summary     => '{{$labels.instance}}: Scrape target is down',
      description => '{{$labels.instance}}: Target down for job {{$labels.job}}',
    }

    prometheus::rule { 'ContainerScrapeError':
      expr        => '(container_scrape_error) != 0',
      for         => '2m',
      summary     => '{{$labels.instance}}: Container scrape error',
      description => '{{$labels.instance}}: Failed to scrape container, metrics will not be updated',
    }
  }
}
