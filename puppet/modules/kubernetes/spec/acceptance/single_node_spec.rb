

require 'securerandom'
require 'spec_helper_acceptance'

describe '::kubernetes' do
  let :cluster_name do
    'test'
  end

  let :namespace do
    "test-#{SecureRandom.hex}"
  end

  let :kubernetes_version do
    if not ENV['KUBERNETES_VERSION'].nil?
      "version => '#{ENV['KUBERNETES_VERSION']}',"
    else
      ''
    end
  end

  context 'test master and worker on a single node, no tls' do
    let :pp do
      "
class{'kubernetes':
  cluster_name                 => '#{cluster_name}',
  service_account_key_generate => true,
  apiserver_insecure_port => 8080,
  #{kubernetes_version}
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
        on host, "ip addr add #{ip}/24 dev eth1 || true"

        # clear firewall
        on host, "iptables -F INPUT"

        # setup kubectl in path with correct kubeconfig
        on host, "echo -e '#!/bin/sh\nKUBECONFIG=/etc/kubernetes/kubeconfig-kubectl exec /opt/bin/kubectl $@' > /usr/bin/kubectl"
        on host, "chmod +x /usr/bin/kubectl"

        # ensure no swap space is mounted
        on host, "swapoff -a"
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
#{pp}

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

    context 'test invalid kubectl::apply manifest, no tls' do
      let :invalid_manifest_apply_pp do
"
#{pp}

kubernetes::apply { 'hello-invalid-namespace':
  type      => 'manifests',
  manifests => ['kind; Namespace\napiVersion: v1\nmetadata:\n  name: testing\n  labels:\n    name: testing']
}
"
      end

      it 'should error when it applies a manifest with a syntax error' do
        hosts_as('k8s-master').each do |host|
          apply_manifest_on(host, invalid_manifest_apply_pp, :expect_failures => true)
        end
      end
    end

    context 'test configmap kubectl::apply manifest, no tls' do
      let :configmap_manifest_apply_pp do
"
#{pp}

kubernetes::apply { 'configmap':
  type      => 'manifests',
  manifests => ['kind: ConfigMap\napiVersion: v1\nmetadata:\n  name: configmap-test\n  labels:\n    name: configmap-test\ndata:\n  example.hello: world']
}
"
      end

      it 'should successfully apply configmap' do
        hosts_as('k8s-master').each do |host|
          apply_manifest_on(host, configmap_manifest_apply_pp, :catch_failures => true)
          expect(
            apply_manifest_on(host, configmap_manifest_apply_pp, :catch_failures => true).exit_code
          ).to be_zero
        end
      end

      it 'should have configured a configmap correctly' do
        result = shell('/opt/bin/kubectl -n kube-system get configmap')
        logger.notify "kubectl -n kube-system get configmap:\n#{result.stdout}"
        expect(result.exit_code).to eq(0)
        expect(result.stdout.scan(/configmap-test/m).size).to eq(1)
        shell("/opt/bin/kubectl -n kube-system delete configmap configmap-test")
      end
    end

    context 'test kubectl::apply_fragment manifest, no tls' do
      let :fragment_apply_pp do
"
#{pp}

kubernetes::apply { 'hello2':
  type      => 'concat',
}

kubernetes::apply_fragment { 'hello2-kind':
  content => 'kind: Namespace',
  order   => '00',
  target  => 'hello2',
}

kubernetes::apply_fragment { 'hello2-apiVersion':
  content => 'apiVersion: v1',
  order   => '01',
  target  => 'hello2',
}

kubernetes::apply_fragment { 'hello2-metadata':
  content => 'metadata:',
  order   => '02',
  target  => 'hello2',
}

kubernetes::apply_fragment { 'hello2-metadata-name':
  content => '  name: testing2',
  order   => '03',
  target  => 'hello2',
}

kubernetes::apply_fragment { 'hello2-metadata-label':
  content => '  labels:',
  order   => '04',
  target  => 'hello2',
}

kubernetes::apply_fragment { 'hello2-metadata-labelname':
  content => '    name: testing2',
  order   => '05',
  target  => 'hello2',
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

      it 'should have configured a namespace from fragments correctly' do
        result = shell('/opt/bin/kubectl get namespace testing2')
        logger.notify "kubectl get namespace testing2:\n#{result.stdout}"
        expect(result.exit_code).to eq(0)
        expect(result.stdout.scan(/Active/m).size).to eq(1)
        shell("/opt/bin/kubectl delete namespace testing2")
      end

    end
  end
end
