# class kubernetes::master
class kubernetes::master (
  $disable_kubelet = false,
  $disable_proxy = false,
){
  include ::kubernetes::apiserver
  include ::kubernetes::controller_manager
  include ::kubernetes::scheduler
  include ::kubernetes::dns
  include ::kubernetes::storage_classes
  include ::kubernetes::kubectl
  if ! $disable_kubelet {
    class{'kubernetes::kubelet':
      role => 'master',
    }
  }
  if ! $disable_proxy {
    class{'kubernetes::proxy':}
  }
}
