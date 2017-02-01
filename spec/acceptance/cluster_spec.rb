require 'spec_helper_acceptance'

describe '::puppernetes' do

  context 'test one master, two worker cluster' do
    let :role do
      'unknown'
    end

    let :cluster_name do
      'test'
    end

    let :global_pp do
      "
class{'puppernetes':
  cluster_name  => '#{cluster_name}',
}

class{'vault_client':
  init_token    => 'init-token-etcd',
  init_role     => '#{cluster_name}-#{role}',
  init_policies => ['#{cluster_name}/#{role}'],
  server_url    => 'http://10.123.0.12:8200'
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

      # Ensure vault-dev server is setup
      hosts_as('vault').each do |host|
        host.shell 'ln -sf /etc/puppetlabs/code/modules/vault_client/files/vault-k8s-server.service /etc/systemd/system/vault-k8s-server.service'
        host.shell 'systemctl daemon-reload'
        host.shell 'systemctl start vault-k8s-server.service'
      end
    end

    # TODO: do vault here

    # Make sure etcd is running and setup as expected
    context 'etcd' do
      let :role do
        'etcd'
      end

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
