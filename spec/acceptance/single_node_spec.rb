require 'spec_helper_acceptance'

if hosts.length == 1
  describe '::etcd' do
    context 'test three single node etcd instances' do
      # Using puppet_apply as a helper
      it 'should work with no errors based on the example' do
        pp = <<-EOS
etcd::instance{'k8s-main':
  version => '3.0.15',
}

etcd::instance{'k8s-events':
  version => '3.0.15',
  client_port => 2389,
  peer_port => 2390,
}

etcd::instance{'k8s-overlay':
  version => '2.3.7',
  client_port => 2399,
  peer_port => 2400,
}
        EOS

        # Run it twice and test for idempotency
        apply_manifest(pp, :catch_failures => true)
        expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero
      end
    end
  end
end
