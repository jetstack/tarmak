require 'spec_helper_acceptance'

describe '::puppernetes' do
  context 'test one master, two worker cluster' do
    let :global_pp do
      "
class{'puppernetes':
  cluster_name => 'beaker',
}

class{'vault_client':
  token => 'beaker-token',
}
"
    end

    before(:all) do
      # assign private ip addresses
      hosts.each do |host|
        ip = host.host_hash[:ip]
        on host, "ifconfig enp0s8 #{ip}/16"
        on host, "iptables -F INPUT"
      end
    end

    # TODO: do vault here

    # Make sure etcd is running and setup as expected
    context 'etcd' do
      let :pp do
        global_pp + "\nclass{'puppernetes::etcd':}"
      end

      it 'should work with no errors based on the example' do
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
