#calico params.pp

class calico::params {
  $install_dir = '/opt/cni'
  $config_dir = '/etc/calico'
  $cni_base_dir = '/etc'
  $helper_dir = '/usr/local/sbin'
  $systemd_dir = '/usr/lib/systemd/system'
  $secure_config_dir = '/root'
  $calico_bin_version = '1.5.5'
  $calico_node_version = '1.0.0'
  $calico_cni_version = '0.3.0'
  $cni_download_url = 'https://github.com/containernetworking/cni/releases/download/v#VERSION#/cni-v#VERSION#.tgz'
  $calico_bin_download_url = 'https://github.com/projectcalico/cni-plugin/releases/download/v#VERSION#'
  $calico_node_download_url = 'https://github.com/projectcalico/calico-containers/releases/download/v#VERSION#'
  $etcd_cert_path = '/etc/etcd/ssl'
  $etcd_cert_base_name = 'etcd-overlay'
  $etcd_overlay_port = 2359
  $kubectl_bin = '/usr/bin/kubectl'
  $kubeconfig = '/etc/kubernetes/kubeconfig-kubelet'
}
