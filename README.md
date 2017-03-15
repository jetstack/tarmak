# aws_ebs

#### Table of Contents

1. [Description](#description)
2. [Classes](#classes)
3. [Defined Types](#defined-types)
## Description
This module attaches, formats (if needed) and mounts EBS volumes in AWS. This
base class just makes sure that all the necessary dependencies are met. To
actually attach & mount a volume you have to use the defined type
`aws_ebs::mount`

## Classes

### `aws_ebs`

This module attaches, formats (if needed) and mounts EBS volumes in AWS. This
base class just makes sure that all the necessary dependencies are met. To
actually attach & mount a volume you have to use the defined type
`aws_ebs::mount`

#### Parameters

##### `bin_dir`

* path to the binary directory for helper scripts
* Type: `["String"]`
* Default: `'/opt/bin'`

##### `systemd_dir`

* path to the directory where systemd units should be placed
* Type: `["String"]`
* Default: `'/etc/systemd/system'`

#### Examples

##### Declaring the base class

```
include ::aws_ebs
```
##### Override binary directory (needs to exist)

```
class{'aws_ebs':
  bin_dir => '/usr/local/sbin',
}
```
## DefinedTypes

### `aws_ebs::mount`

This defined type attaches, formats (if needed) and mounts a single EBS
volume in AWS.

#### Parameters

##### `volume_id`

* the volume id of the AWS EBS volume
* Type: `["String"]`

##### `dest_path`

* where to mount the device (needs to exists)
* Type: `["String"]`

##### `device`

* block device to attach to (should be `/dev/xvd[a-z]`)
* Type: `["String"]`

##### `filesystem`

* select the filesystem to initialize a volume
* Type: `["Enum['xfs']"]`
* Default: `'xfs'`

#### Examples

##### Attach, format & mount EBS volume

```
aws_ebs::mount{'data':
  volume_id => 'vol-deadbeef',
  device    => '/dev/xvdd',
  dest_path => '/mnt',
}
```
