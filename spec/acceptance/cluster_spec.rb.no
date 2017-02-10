require 'spec_helper_acceptance'

describe '::puppernetes' do

  context 'test one master, two worker cluster' do
    let :role do
      'unknown'
    end

    let :vault_ip do
        vault_ip = nil
        hosts_as('vault').each do |host|
          vault_ip = host.host_hash[:ip]
        end
        vault_ip
    end

    let :cluster_name do
      'test'
    end

    let :global_pp do
      "
class{'puppernetes':
  cluster_name                  => '#{cluster_name}',
  etcd_instances                => 1,
  etcd_advertise_client_network => '10.0.0.0/8'
}

class{'vault_client':
  init_token    => 'init-token-#{role}',
  init_role     => '#{cluster_name}-#{role}',
  init_policies => ['#{cluster_name}/#{role}'],
  server_url    => 'http://#{vault_ip}:8200'
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
        on host, 'ln -sf /etc/puppetlabs/code/modules/vault_client/files/vault-k8s-server.service /etc/systemd/system/vault-k8s-server.service'
        on host, 'systemctl daemon-reload'
        on host, 'systemctl start vault-k8s-server.service'
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

    context 'master' do
      let :role do
        'master'
      end

      let :pp do
        global_pp + "\nclass{'puppernetes::master':}"
      end

      it 'should work with no errors based on the example' do
        hosts_as('master').each do |host|
          apply_manifest_on(host, pp, :catch_failures => true)
          expect(
            apply_manifest_on(host, pp, :catch_failures => true).exit_code
          ).to be_zero
        end
      end
    end
  end
end
