define calico::ipPool (
  String $ip_pool,
  Integer $ip_mask,
  String $ipip_enabled
)
{
  file { "/etc/calico/ipPool-${ip_pool}.yaml":
    ensure => file,
    content => template('calico/ipPool.yaml.erb'),
  }

  exec { "Configure calico ipPool for CIDR $ip_pool":
    path => [ '/bin', '/usr/bin' ],
    user => "root",
    command => "/bin/bash -c \"`/usr/bin/grep ETCD_ENDPOINTS /etc/calico/calico.env` `/usr/bin/grep ETCD_CERT_FILE /etc/calico/calico.env` `/usr/bin/grep ETCD_KEY_FILE /etc/calico/calico.env` `/usr/bin/grep ETCD_CA_CERT_FILE /etc/calico/calico.env` /opt/cni/bin/calicoctl apply -f /etc/calico/ipPool-${ip_pool}.yaml\"",
    unless => "/bin/bash -c \"`/usr/bin/grep ETCD_ENDPOINTS /etc/calico/calico.env` `/usr/bin/grep ETCD_CERT_FILE /etc/calico/calico.env` `/usr/bin/grep ETCD_KEY_FILE /etc/calico/calico.env` `/usr/bin/grep ETCD_CA_CERT_FILE /etc/calico/calico.env` /opt/cni/bin/calicoctl get -f /etc/calico/ipPool-${ip_pool}.yaml | /usr/bin/grep ${ip_pool}/${ip_mask}\"",
    require => [ Service["calico-node"], File["/opt/cni/bin/calicoctl"], File["/etc/calico/ipPool-${ip_pool}.yaml"], Exec["Trigger etcd overlay cert"] ],
  }
}
