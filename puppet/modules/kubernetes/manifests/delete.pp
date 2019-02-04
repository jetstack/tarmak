# adds resources to a kubernetes master
define kubernetes::delete(
  $format = 'yaml',
){
  require ::kubernetes

  $apply_file = "${::kubernetes::apply_dir}/${name}.${format}"

  file{$apply_file:
    ensure  => absent,
  }
}
