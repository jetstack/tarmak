

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
  cluster_name                 => '#{cluster_name}',
  service_account_key_generate => true,
}
class{'kubernetes::master':
  disable_kubelet => true,
  disable_proxy => true,
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

    context 'test kubectl::apply manifest, no tls' do
      let :manifest_apply_pp do
"
include kubernetes::apiserver

kubernetes::apply { 'hello':
  type      => 'manifests',
  manifests => ['kind: Namespace\napiVersion: v1\nmetadata:\n  name: testing\n  labels:\n    name: testing']
}
"
      end

      it 'should apply a manifest correctly' do
        hosts_as('k8s-master').each do |host|
          apply_manifest_on(host, manifest_apply_pp, :catch_failures => true)
          expect(
            apply_manifest_on(host, manifest_apply_pp, :catch_failures => true).exit_code
          ).to be_zero
        end
      end

      it 'should have a testing namespace' do
        result = shell('/opt/bin/kubectl get namespace testing')
        logger.notify "kubectl get namespace testing:\n#{result.stdout}"
        expect(result.exit_code).to eq(0)
        expect(result.stdout.scan(/Active/m).size).to eq(1)
        shell("/opt/bin/kubectl delete namespace testing")
      end
    end

    context 'test kubectl::apply_fragment manifest, no tls' do
      let :fragment_apply_pp do
"
include kubernetes::apiserver

kubernetes::apply { 'hello':
  type      => 'concat',
}

kubernetes::apply_fragment { 'hello-world-kind':
  content => ['kind: Namespace'],
  order   => '00',
}

kubernetes::apply_fragment { 'hello-world-apiVersion':
  content => 'apiVersion: v1',
  order   => '01',
}

kubernetes::apply_fragment { 'hello-world-metadata':
  content => 'metadata:',
  order   => '02',
}

kubernetes::apply_fragment { 'hello-world-metadata-name':
  content => '  name: testing2',
  order   => '03',
}

kubernetes::apply_fragment { 'hello-world-metadata-label':
  content => '  labels:',
  order   => '04',
}

kubernetes::apply_fragment { 'hello-world-metadata-labelname':
  content => '    name: testing2',
  order   => '05',
}
"
      end

      it 'should apply a fragment manifest correctly' do
        hosts_as('k8s-master').each do |host|
          apply_manifest_on(host, fragment_apply_pp, :catch_failures => true)
          expect(
            apply_manifest_on(host, fragment_apply_pp, :catch_failures => true).exit_code
          ).to be_zero
        end
      end

    end
  end
end
