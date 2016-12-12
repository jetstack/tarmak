require 'spec_helper_acceptance'

if hosts.length == 3
  describe '::etcd' do
    context 'test three three-node etcd clusters' do

      before(:all) do
        # assign private ip addresses
        hosts_as('etcd').each do |host|
          ip = host.host_hash[:ip]
          on host, "ifconfig enp0s8 #{ip}/24"
          on host, "date > /tmp/test"
        end
      end

      # Using puppet_apply as a helper
      it 'should work with no errors based on the example' do
        pp = <<-EOS
$members = 3
$initial_cluster = ['etcd1','etcd2','etcd3']

etcd::instance{'k8s-main':
  version         => '3.0.15',
  members         => $members,
  initial_cluster => $initial_cluster,
}

etcd::instance{'k8s-events':
  version         => '3.0.15',
  members         => $members,
  initial_cluster => $initial_cluster,
  client_port     => 2389,
  peer_port       => 2390,
}

etcd::instance{'k8s-overlay':
  version         => '2.3.7',
  members         => $members,
  initial_cluster => $initial_cluster,
  client_port     => 2399,
  peer_port       => 2400,
}
        EOS

        hosts_as('etcd').each do |host|
          apply_manifest_on(host, pp, :catch_failures => true)
          expect(
            apply_manifest_on(host, pp, :catch_failures => true).exit_code
          ).to be_zero
        end
      end
    end
  end
end
