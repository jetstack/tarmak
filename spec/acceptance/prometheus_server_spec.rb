require 'securerandom'
require 'spec_helper_acceptance'

describe '::prometheus::server' do

  context 'prometheus server' do
    let :pp do
      "
class{'prometheus::server':
}
class{'prometheus::node_exporter':
}
class{'prometheus::kube_state_metrics':
}
"
    end

    before(:all) do
      hosts.each do |host|
        # reset firewall
        on host, "iptables -F INPUT"

        # make sure docker + curl is installed
        if fact_on(host, 'osfamily') == 'RedHat'
          on(host, 'yum install -y docker curl')
        end

        # start docker
        on host, 'systemctl start docker.service'

        # setup minikube
        on host, 'ln -sf /etc/puppetlabs/code/modules/prometheus/files/minikube.service /etc/systemd/system/minikube.service'
        on host, 'ln -sf /etc/puppetlabs/code/modules/prometheus/files/localkube.service.d /etc/systemd/system/localkube.service.d'
        on host, 'systemctl daemon-reload'
        on host, 'systemctl start minikube.service'

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
