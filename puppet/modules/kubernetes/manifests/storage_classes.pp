# This class sets up the default storage classes for cloud providers
class kubernetes::storage_classes{
  $cloud_provider = $::kubernetes::cloud_provider
  if versioncmp($::kubernetes::version, '1.4.0') >= 0 {
    if $cloud_provider == 'aws' {
      kubernetes::apply{'storage-classes':
        manifests => [
          template('kubernetes/storage-classes-aws.yaml.erb'),
        ],
      }
    }
  }
}
