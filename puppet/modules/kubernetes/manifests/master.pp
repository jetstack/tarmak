# class kubernetes::master
class kubernetes::master (
  $disable_kubelet = false,
  $disable_proxy = false,
){
  include ::kubernetes::apiserver
  include ::kubernetes::rbac
  include ::kubernetes::controller_manager
  include ::kubernetes::scheduler
  include ::kubernetes::dns
  include ::kubernetes::storage_classes
  include ::kubernetes::kubectl
  include ::kubernetes::pod_security_policy
  if ! $disable_kubelet {
    class{'kubernetes::kubelet':
      role => 'master',
    }
  }
  if ! $disable_proxy {
    class{'kubernetes::proxy':}
  }
}
