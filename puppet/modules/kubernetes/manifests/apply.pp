# adds resources to a kubernetes master
define kubernetes::apply(
  $manifests = [],
  $force = false,
  $format = 'yaml',
  $systemd_wants = [],
  $systemd_requires = [],
  $systemd_after = [],
  $systemd_before = [],
  Enum['manifests','concat'] $type = 'manifests',
){
  require ::kubernetes
  require ::kubernetes::addon_manager

  if ! defined(Class['kubernetes::apiserver']) {
    fail('This defined type can only be used on the kubernetes master')
  }

  $manifests_content = $manifests.join("\n---\n")
  $apply_file = "${::kubernetes::apply_dir}/${name}.${format}"

  case $type {
    'manifests': {
      file{$apply_file:
        ensure  => file,
        mode    => '0640',
        owner   => 'root',
        group   => $kubernetes::group,
        content => $manifests_content,
      }
    }
    'concat': {
      concat { $apply_file:
        ensure         => present,
        ensure_newline => true,
        mode           => '0640',
        owner          => 'root',
        group          => $kubernetes::group,
      }
    }
    default: {
      fail("Unknown type parameter: '${type}'")
    }
  }
}
