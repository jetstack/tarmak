

require 'securerandom'
require 'spec_helper_acceptance'

describe '::kubernetes' do
  let :cluster_name do
    'test'
  end

  let :namespace do
    "test-#{SecureRandom.hex}"
  end

  context 'test master and worker on a single node, no tls' do
    let :pp do
      "
class{'kubernetes':
  cluster_name => '#{cluster_name}',
}
class{'kubernetes::master':
  disable_kubelet => true,
}
class{'kubernetes::worker':
}
      "
    end

    before(:all) do
      # assign private ip addresses
      hosts.each do |host|
        ip = host.host_hash[:ip]
        on host, "ifconfig enp0s8 #{ip}/24"
        on host, "iptables -F INPUT"
      end

      # Ensure etcd is setup
      hosts_as('master').each do |host|
        on host, "ln -sf #{$module_path}kubernetes/files/etcd.service /etc/systemd/system/etcd.service"
        on host, 'systemctl daemon-reload'
        on host, 'systemctl start etcd.service'
      end

      # Ensure docker is setup
      hosts_as('master').each do |host|
        on host, 'yum install -y docker'
        on host, 'systemctl start docker.service'
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

    it 'should setup a healthy master' do
      # verify master setup
      result = shell('/opt/bin/kubectl get cs')
      logger.notify "kubectl get cs:\n#{result.stdout}"
      expect(result.exit_code).to eq(0)
      expect(result.stdout.scan(/Healthy/m).size).to eq(3)
    end

    it 'should have a ready node' do
      result = shell('/opt/bin/kubectl get nodes')
      logger.notify "kubectl get nodes:\n#{result.stdout}"
      expect(result.exit_code).to eq(0)
      expect(result.stdout.scan(/Ready/m).size).to eq(1)
    end

    it 'should pass a smoke test' do
      skip "not ready yet"
      shell("/opt/bin/kubectl create ns #{namespace}")
      shell("/opt/bin/kubectl kubectl run --namespace=#{namespace} nginx --replicas=2 --image=nginx")
      shell("/opt/bin/kubectl kubectl expose --namespace=#{namespace} deployment nginx --port=80")
      shell("/opt/bin/kubectl delete ns #{namespace}")
    end
  end
end
