# Concat fragment for apply
define kubernetes::apply_fragment(
  $content,
  $order,
  $format = 'yaml',
){
  require ::kubernetes
  require ::kubernetes::kubectl

  if ! defined(Class['kubernetes::apiserver']) {
    fail('This defined type can only be used on the kubernetes master')
  }

  $apply_file = "${::kubernetes::apply_dir}/${name}.${format}"

  concat::fragment { "kubectl-apply-${name}":
    target  => $apply_file,
    content => $content,
    order   => $order,
  }
}
