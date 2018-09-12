require 'spec_helper_acceptance'

if hosts.length == 1
  describe '::etcd' do
    context 'test three single node etcd instances' do
      # Using puppet_apply as a helper
      it 'should work with no errors based on the example' do
        pp = <<-EOS
$advertise_client_network = '10.0.0.0/8'

etcd::instance{'k8s-main':
  version                  => '3.2.24',
  advertise_client_network => $advertise_client_network,
}

etcd::instance{'k8s-events':
  version                  => '3.2.24',
  client_port              => 2389,
  peer_port                => 2390,
  advertise_client_network => $advertise_client_network,
}

etcd::instance{'k8s-overlay':
  version                  => '3.2.24',
  client_port              => 2399,
  peer_port                => 2400,
  advertise_client_network => $advertise_client_network,
}
        EOS

        # Run it twice and test for idempotency
        apply_manifest(pp, :catch_failures => true)
        expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero
      end

      [2379, 2389, 2399].each do |port|
        it "test etcd on port #{port} on host #{host.name}" do
				result = host.shell "ETCDCTL=http://127.0.0.1:#{port} /opt/etcd-3.2.24/etcdctl cluster-health"
        end
      end
    end
  end
end
