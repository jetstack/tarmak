require 'securerandom'
require 'spec_helper_acceptance'

$ip = '10.0.2.15'

describe '::pupperentes::single_node' do
  let :cluster_name do
    'test'
  end

  let :namespace do
    "test-#{SecureRandom.hex}"
  end

  let :kubernetes_version do
    ENV['KUBERNETES_VERSION'] || '1.5.7'
  end

  let :kubernetes_authorization_mode do
    ENV['KUBERNETES_AUTHORIZATION_MODE'] || '[]'
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

class{'puppernetes::single_node':
  cluster_name                  => '#{cluster_name}',
  etcd_advertise_client_network => '10.0.0.0/8',
  kubernetes_api_url            => 'https://#{$ip}:6443',
  kubernetes_version            => '#{kubernetes_version}',
  kubernetes_authorization_mode => #{kubernetes_authorization_mode},
}


"
    end

    before(:all) do
      hosts.each do |host|
        # reset firewall
        on host, "iptables -F INPUT"

        # make hostname resolvable
        line = "#{$ip} k8s.test.jetstack.net k8s"
        on(host, "grep -q \"#{line}\" /etc/hosts || echo \"#{line}\" >> /etc/hosts")

        # make sure curl unzip vim is installed
        if fact_on(host, 'osfamily') == 'RedHat'
          on(host, 'yum install -y unzip docker')
        elsif fact_on(host, 'osfamily') == 'Debian'
          on(host, 'apt-get install -y unzip apt-transport-https ca-certificates curl python-software-properties')
          on(host, 'apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D')
          on(host, 'echo "deb https://apt.dockerproject.org/repo debian-jessie main" > /etc/apt/sources.list.d/docker.list')
          on(host, 'apt-get update')
          on(host, 'apt-get -y install docker-engine')
        end

        # setup develop vault server
        on host, 'ln -sf /etc/puppetlabs/code/modules/vault_client/files/vault-k8s-server.service /etc/systemd/system/vault-k8s-server.service'
        on host, 'systemctl daemon-reload'
        on host, 'systemctl start vault-k8s-server.service'

        # start docker
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

    it 'should setup a healthy master', :retry => 20, :retry_wait => 5 do
      # verify master setup
      result = shell('/opt/bin/kubectl version')
      logger.notify "kubectl version:\n#{result.stdout}"
      expect(result.exit_code).to eq(0)
      result = shell('/opt/bin/kubectl get cs')
      logger.notify "kubectl get cs:\n#{result.stdout}"
      expect(result.exit_code).to eq(0)
      expect(result.stdout.scan(/Healthy/m).size).to be >= 2
    end

    it 'should have a ready node', :retry => 20, :retry_wait => 5 do
      result = shell('/opt/bin/kubectl get nodes')
      logger.notify "kubectl get nodes:\n#{result.stdout}"
      expect(result.exit_code).to eq(0)
      expect(result.stdout.scan(/Ready/m).size).to eq(1)
    end

    it 'should have three ready dns pods', :retry => 20, :retry_wait => 5 do
      result = shell('/opt/bin/kubectl get pods --namespace kube-system -l k8s-app=kube-dns')
      logger.notify "kubectl get pods:\n#{result.stdout}"
      expect(result.exit_code).to eq(0)
      expect(result.stdout.scan(/Running/m).size).to eq(3)
      expect(result.stdout.scan(/(4\/4|3\/3)/m).size).to eq(3)
    end

    it 'should have a ready dns autoscaler pod', :retry => 20, :retry_wait => 5 do
      result = shell('/opt/bin/kubectl get pods --namespace kube-system -l k8s-app=kube-dns-autoscaler')
      logger.notify "kubectl get pods:\n#{result.stdout}"
      expect(result.exit_code).to eq(0)
      expect(result.stdout.scan(/Running/m).size).to eq(1)
      expect(result.stdout.scan(/1\/1/m).size).to eq(1)
    end
  end
end
