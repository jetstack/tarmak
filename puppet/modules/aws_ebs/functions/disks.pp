function aws_ebs::disks() >> Array[String] {
  if $::disks == undef {
    []
  } else {
    $::disks.keys.sort
  }
}
