require 'securerandom'
require 'spec_helper_acceptance'

$ip = '10.0.2.15'

describe '::puppetpacker' do
  let :cluster_name do
    'test'
  end

  let :kubernetes_version do
    ENV['KUBERNETES_VERSION'] || '1.9.7'
  end

  context 'Packer install of Tarmak' do
    let :cluster_name do
      'test'
    end

    let :pp do
      "
class{'tarmak':
  service_ensure                => 'stopped',
  kubernetes_version            => '#{kubernetes_version}',
}
    
class{'vault_client':
  init_token => 'init-token-all',
  init_role => 'test-all',
  server_url => 'http://127.0.0.1:8200',
  run_exec      => false,
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
          on(host, 'cp -a /usr/lib/systemd/system/docker.service /etc/systemd/system/docker.service')
          on(host, 'sed -i -e \'s/systemd/cgroupfs/g\' /etc/systemd/system/docker.service')
          on(host, 'systemctl daemon-reload')
        elsif fact_on(host, 'osfamily') == 'Debian'
          on(host, 'apt-get install -y unzip apt-transport-https ca-certificates curl python-software-properties')
          on(host, 'apt-key add /etc/puppetlabs/code/modules/tarmak/spec/files/ubuntu-16-04-docker.gpg')
          on(host, 'echo "deb https://apt.dockerproject.org/repo debian-jessie main" > /etc/apt/sources.list.d/docker.list')
          on(host, 'apt-get update')
          on(host, 'apt-get -y install docker-engine')
        end

        # reset firewall
        on host, "iptables -F INPUT"

        # ensure no swap space is mounted
        on host, "swapoff -a"

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
  end
end
