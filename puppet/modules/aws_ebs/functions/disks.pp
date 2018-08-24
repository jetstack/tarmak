function aws_ebs::disks() >> Array[String] {
  # if disks fact is undefined then return empty array
  if $::disks == undef {
    []
  } else {
    # For each disk, we split by continuous non-number characters by Regexp.
    # This creates an array of array strings of numbers of each device name.
    # We then convert these strings into numbers. For first elements with no
    # number in (devices starting with letter) that contain '' we just assign
    # that element 0. These number arrays representing the devices are then
    # zipped onto the devices which can then be correctly sorted. Finally get
    # the device name of each tuple to return as correctly sorted.

    $::disks.keys.map |$device| {
      split($device, Regexp['[^\d]+']).map |$number_in_device_name| {
        if $number_in_device_name == '' {
          0
        } else {
          $number_in_device_name + 0
        }
      }
    }.zip($::disks.keys).sort.map |$i| { $i[1] }
  }
}
