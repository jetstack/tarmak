# Install/configure an consul node.
#
# @param data_dir The directory to store consul data
# @param config_dir The directory to store consul config
# @param user The username to run consul
# @param uid The user ID to run consul
# @param group The group to run consul
# @param gid The consul group ID
# @param version Version of consul to deploy
# @param cloud_provider Select cloud provider for consul discovery
# @param exporter_enabled Enable/disable prometheus exporter
# @param exporter_version Version of prometheus exporter
# @param backup_enabled Enable/disable backup
# @param backup_version Version of backup
# @param advertise_network Specify network used for consul
class consul(
    String $data_dir = '/var/lib/consul',
    String $config_dir = '/etc/consul',
    String $dest_dir = '/opt',
    Integer $uid = 871,
    Integer $gid = 871,
    String $user = 'consul',
    String $group = 'consul',
    String $version = '1.0.3',
    Enum['aws', ''] $cloud_provider = '',
    Boolean $exporter_enabled = true,
    String $exporter_version = '0.3.0',
    Boolean $backup_enabled = true,
    String $backup_version = 'xx',
    String $acl_default_policy = 'deny',
    String $acl_down_policy = 'deny',
    Boolean $server = true,
    String $client_addr = '127.0.0.1',
    String $bind_addr = '0.0.0.0',
    String $log_level = 'INFO',
    String $datacenter = 'dc1',
    Optional[String] $advertise_network = undef,
    Optional[Array[String]] $retry_join = undef,
    Optional[String] $ca_file = undef,
    Optional[String] $cert_file = undef,
    Optional[String] $key_file = undef,
    $consul_encrypt = true,
    $private_ip,
    $consul_master_token,
    $region,
    $instance_count,
    $environment,
) inherits ::consul::params {

    include ::archive
    include ::airworthy

    $app_name = 'consul'
    $_dest_dir = "${dest_dir}/${app_name}-${version}"
    $bin_path = "${_dest_dir}/${app_name}"
    $link_path = "${dest_dir}/bin/${app_name}"
    $config_path = "${config_dir}/consul.json"

    $exporter_dest_dir = "${dest_dir}/${app_name}_exporter-${exporter_version}"
    $exporter_bin_path = "${exporter_dest_dir}/${app_name}_exporter"

    Class['::airworthy']
    -> class { '::consul::install': }
    -> class { '::consul::config': }
    ~> class { '::consul::service': }
}
