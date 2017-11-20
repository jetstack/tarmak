# This class disable the source/destination check on AWS instances
class calico::disable_source_destination_check(
  String $image = 'ottoyiu/k8s-ec2-srcdst',
  String $version = '0.1.0',
){
  include calico

  if $calico::cloud_provider != 'aws' {
    fail('This class is meant to be used with cloud_provider AWS only')
  }

  # ensure old systemd service gets removed
  $service_name = 'disable-source-destination-check.service'
  $bin_path = "${::calico::bin_dir}/disable-source-destination-check.sh"
  file {$bin_path:
    ensure  => absent,
  }
  file {"${::calico::systemd_dir}/${service_name}":
    ensure  => absent,
  }

  # deploy service on masters only
  if defined(Class['kubernetes::apiserver']) {
    include ::kubernetes

    $namespace = $::calico::namespace
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

    $aws_region = $::ec2_metadata['placement']['availability-zone'][0,-2]

    kubernetes::apply{'disable-srcdest-node':
      manifests => [
        template('calico/disable-source-destination.yaml.erb'),
      ],
    }
  }
}
