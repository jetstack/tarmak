# Class: airworthy
class airworthy (
  String $version = '0.2.0',
  String $checksum_type = 'sha256',
  Optional[String] $checksum = undef,
  String $download_url = 'https://github.com/jetstack/airworthy/releases/download/#VERSION#/airworthy_#VERSION#_linux_amd64',
  String $dest_dir = '/opt',
  String $bin_dir = '/opt/bin',
){

  $_dest_dir = "${dest_dir}/airworthy-${version}"

  $_download_url = regsubst(
    $download_url,
    '#VERSION#',
    $version,
    'G'
  )

  if $checksum == undef {
    $_checksum = $checksum_type ? {
      'sha256' => $version ? {
        '0.2.0'  => '2d69cfe0b92f86481805c28d0b8ae47a8ffa6bb2373217e7c5215d61fc9efa1d',
        default  => undef,
      },
      default  => undef,
    }
  } else {
      $_checksum = $checksum
  }
}
