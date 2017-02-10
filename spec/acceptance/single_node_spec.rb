require 'securerandom'
require 'spec_helper_acceptance'

describe '::pupperentes::single_node' do
  let :cluster_name do
    'test'
  end

  let :namespace do
    "test-#{SecureRandom.hex}"
  end

  context 'single node with master + worker components' do
    let :cluster_name do
      'test'
    end

    let :pp do
      "
class{'vault_client':
  token      => 'root-token',
  server_url => 'http://127.0.0.1:8200',
}

class{'puppernetes':
  cluster_name                  => '#{cluster_name}',
  etcd_instances                => 1,
  etcd_advertise_client_network => '10.0.0.0/8',
  kubernetes_api_url => 'https://10.0.2.15:6443'
}

class{'puppernetes::single_node':}

"
    end

    before(:all) do
      hosts.each do |host|
        # reset firewall
        on host, "iptables -F INPUT"

        # setup develop vault server
        on host, 'ln -sf /etc/puppetlabs/code/modules/vault_client/files/vault-k8s-server.service /etc/systemd/system/vault-k8s-server.service'
        on host, 'systemctl daemon-reload'
        on host, 'systemctl start vault-k8s-server.service'

        # setup docker
        on host, 'yum install -y docker'
        on host, 'systemctl start docker.service'
      end
    end

    it 'should converge on the first puppet run' do
      hosts.each do |host|
        apply_manifest_on(host, pp, :catch_failures => true)
        expect(
          apply_manifest_on(host, pp, :catch_failures => true).exit_code
        ).to be_zero
      end
    end
  end
end
