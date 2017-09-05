require 'spec_helper_acceptance'

if hosts.length == 3
  describe '::etcd' do
    context 'test three three-node etcd clusters' do

      before(:all) do
        # assign private ip addresses
        hosts_as('etcd').each do |host|
          ip = host.host_hash[:ip]
          on host, "ip addr add #{ip}/16 dev eth1"
          on host, "iptables -F INPUT"
        end
      end

      # Using puppet_apply as a helper
      it 'should work with no errors based on the example' do
        pp = <<-EOS
$members = 3
$initial_cluster = ['etcd1','etcd2','etcd3']
$advertise_client_network = '10.123.0.0/16'

etcd::instance{'k8s-main':
  version                  => '3.0.15',
  nodename                 => $::hostname,
  members                  => $members,
  initial_cluster          => $initial_cluster,
  advertise_client_network => $advertise_client_network,
}

etcd::instance{'k8s-events':
  version                  => '3.0.15',
  nodename                 => $::hostname,
  members                  => $members,
  initial_cluster          => $initial_cluster,
  advertise_client_network => $advertise_client_network,
  client_port              => 2389,
  peer_port                => 2390,
}

etcd::instance{'k8s-overlay':
  version                  => '2.3.7',
  nodename                 => $::hostname,
  members                  => $members,
  initial_cluster          => $initial_cluster,
  advertise_client_network => $advertise_client_network,
  client_port              => 2399,
  peer_port                => 2400,
}
        EOS

        threads = []

        hosts_as('etcd').each do |host|
          # run apply in parallel
          threads << Thread.new do
            apply_manifest_on(host, pp, :catch_failures => true)
            expect(
              apply_manifest_on(host, pp, :catch_failures => true).exit_code
            ).to be_zero
          end
        end

        # wait for all nodes to be applied
        threads.each do |thr|
          thr.join
        end
      end

      [2379, 2389, 2399].each do |port|
        hosts_as('etcd').each do |host|
          it "test etcd on port #{port} on host #{host.name}" do
            result = host.shell "ETCDCTL=http://127.0.0.1:#{port} /opt/etcd-3.0.15/etcdctl cluster-health"
          end
        end
      end
    end
  end
end
