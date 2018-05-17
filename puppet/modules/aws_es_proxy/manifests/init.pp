class aws_es_proxy (
  $version = $::aws_es_proxy::params::version,
  $download_url = $::aws_es_proxy::params::download_url,
  $dest_dir = $::aws_es_proxy::params::dest_dir,
) inherits ::aws_es_proxy::params {

  $path = defined('$::path') ? {
    default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
    true    => $::path,
  }

  $_dest_dir = "${dest_dir}/aws-es-proxy-${version}"
  $_download_url = regsubst(
    $download_url,
    '#VERSION#',
    $version,
    'G'
  )

  $proxy_path = "${_dest_dir}/aws-es-proxy"

  class { '::aws_es_proxy::install': }
  -> Class['::aws_es_proxy']

}
