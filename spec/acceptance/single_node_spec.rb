require 'spec_helper_acceptance'

describe '::kubernetes' do
  let :cluster_name do
    'test'
  end

  context 'test master and worker on a single node, no tls' do
    let :pp do
      "
class{'kubernetes':
  cluster_name => '#{cluster_name}',
}
class{'kubernetes::master':}
      "
    end

    before(:all) do
      # assign private ip addresses
      hosts.each do |host|
        ip = host.host_hash[:ip]
        on host, "ifconfig enp0s8 #{ip}/24"
        on host, "iptables -F INPUT"
      end

      # Ensure vault-dev server is setup
      hosts_as('master').each do |host|
        on host, "ln -sf #{$module_path}kubernetes/files/etcd.service /etc/systemd/system/etcd.service"
        on host, 'systemctl daemon-reload'
        on host, 'systemctl start etcd.service'
      end
    end
    it 'should setup single node without errors based on the example' do
      hosts_as('k8s-master').each do |host|
        apply_manifest_on(host, pp, :catch_failures => true)
        expect(
          apply_manifest_on(host, pp, :catch_failures => true).exit_code
        ).to be_zero
      end
    end
  end
end
