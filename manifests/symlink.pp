# adds a symlink to hyperkube
define kubernetes::symlink (
){
  include kubernetes::install
  File["${::kubernetes::_dest_dir}/hyperkube"] ->
  file { "${::kubernetes::_dest_dir}/${title}":
    ensure => link,
    target => 'hyperkube',
  }

  ensure_resource('file', [$::kubernetes::bin_dir], {
    ensure => directory,
    mode   => '0755',
  })

  File[$::kubernetes::bin_dir] ->
  file { "${::kubernetes::bin_dir}/${title}":
    ensure => link,
    target => "${::kubernetes::_dest_dir}/hyperkube",
  }
}
