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
    ENV['KUBERNETES_VERSION'] || '1.8.10'
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
  init_token => 'init-token-all',
  init_role => 'test-all',
  server_url => 'http://127.0.0.1:8200',
}

class{'tarmak::single_node':
  cluster_name                  => '#{cluster_name}',
  etcd_advertise_client_network => '10.0.0.0/8',
  kubernetes_api_url            => 'https://api.test.jetstack.net:6443',
  kubernetes_version            => '#{kubernetes_version}',
  kubernetes_authorization_mode => #{kubernetes_authorization_mode},
}


"
    end

    before(:all) do
      hosts.each do |host|
        # make hostname resolvable
        line = "#{host.host_hash[:ip]} k8s.test.jetstack.net api.test.jetstack.net k8s"
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

        # reset firewall
        on host, "iptables -F INPUT"

        # setup kubectl in path with correct kubeconfig
        on host, "echo -e '#!/bin/sh\nKUBECONFIG=/etc/kubernetes/kubeconfig-kubectl exec /opt/bin/kubectl $@' > /usr/bin/kubectl"
        on host, "chmod +x /usr/bin/kubectl"

        # ensure no swap space is mounted
        on host, "swapoff -a"

        # setup develop vault server
        on host, 'ln -sf /etc/puppetlabs/code/modules/vault_client/files/vault-dev-server.service /etc/systemd/system/vault-dev-server.service'
        on host, 'systemctl daemon-reload'
        on host, 'systemctl start vault-dev-server.service'

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
      result = shell('kubectl version')
      logger.notify "kubectl version:\n#{result.stdout}"
      expect(result.exit_code).to eq(0)
      result = shell('kubectl get cs')
      logger.notify "kubectl get cs:\n#{result.stdout}"
      expect(result.exit_code).to eq(0)
      expect(result.stdout.scan(/Healthy/m).size).to be >= 2
    end

    it 'should have a ready node', :retry => 20, :retry_wait => 5 do
      result = shell('kubectl get nodes')
      logger.notify "kubectl get nodes:\n#{result.stdout}"
      expect(result.exit_code).to eq(0)
      expect(result.stdout.scan(/Ready/m).size).to eq(1)
    end

    it 'should have three ready dns pods', :retry => 20, :retry_wait => 5 do
      result = shell('kubectl get pods --namespace kube-system -l k8s-app=kube-dns')
      logger.notify "kubectl get pods:\n#{result.stdout}"
      expect(result.exit_code).to eq(0)
      expect(result.stdout.scan(/Running/m).size).to eq(3)
      expect(result.stdout.scan(/(4\/4|3\/3)/m).size).to eq(3)
    end

    it 'should have a ready dns autoscaler pod', :retry => 20, :retry_wait => 5 do
      result = shell('kubectl get pods --namespace kube-system -l k8s-app=kube-dns-autoscaler')
      logger.notify "kubectl get pods:\n#{result.stdout}"
      expect(result.exit_code).to eq(0)
      expect(result.stdout.scan(/Running/m).size).to eq(1)
      expect(result.stdout.scan(/1\/1/m).size).to eq(1)
    end

    context 'with podsecuritypolicy enabled' do
      before(:all) do
        begin
          shell('grep PodSecurityPolicy /etc/systemd/system/kube-apiserver.service')
        rescue Beaker::Host::CommandFailure
          skip('PodSecurityPoliciy is not enabled')
        end

        result = shell('kubectl create namespace developer')
        expect(result.exit_code).to eq(0)

        result = shell('kubectl create rolebinding developer-admin-binding --clusterrole=admin --user=developer --namespace=developer')
        expect(result.exit_code).to eq(0)

        result = shell('kubectl create rolebinding kubeadmin-admin-binding --clusterrole=admin --user=kubeadmin --namespace=kube-system')
        expect(result.exit_code).to eq(0)
      end

      after(:all) do
        begin
          shell('kubectl delete namespace developer')
        rescue Beaker::Host::CommandFailure
        end

        begin
          shell('kubectl delete rolebinding kubeadmin-admin-binding -n kube-system')
        rescue Beaker::Host::CommandFailure
        end
      end

      it 'allows developer to run unprivileged pods in namespace developer' do
        result = shell('kubectl --as=developer run busybox --image=busybox --restart=Never -n developer --rm --attach  -- uname -a')
        expect(result.exit_code).to eq(0)
      end

      it 'forbids developer to run privileged pods in namespace developer' do
        if Gem::Version.new(kubernetes_version) >= Gem::Version.new('1.10.0') and Gem::Version.new(kubernetes_version) < Gem::Version.new('1.11.0')
          skip('issue #61713 prevents this test from working with 1.10')
        end

        expect {
          shell('kubectl --as=developer run busybox-priv --image=busybox --restart=Never -n developer --overrides \'{"spec":{"containers":[{"name":"busybox","image":"busybox","command":["uname","-a"],"securityContext":{"privileged":true}}]}}\' --rm --attach')
        }.to raise_error(Beaker::Host::CommandFailure, /forbidden/)
      end

      it 'allows kubeadmins to run privileged pods in namespace kube-system' do
        result = shell('kubectl --as=kubeadmin run busybox-priv --image=busybox --restart=Never -n kube-system --overrides \'{"spec":{"containers":[{"name":"busybox","image":"busybox","command":["uname","-a"],"securityContext":{"privileged":true}}]}}\' --rm --attach')
        expect(result.exit_code).to eq(0)
      end
    end
  end
end
