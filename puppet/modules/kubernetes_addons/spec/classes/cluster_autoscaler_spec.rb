require 'spec_helper'
describe 'kubernetes_addons::cluster_autoscaler' do
  let(:facts) do
    {
      :ec2_metadata => {
        'placement' => {
          'availability-zone' => 'eu-west-1a',
        }
      }
    }
  end

  let(:kubernetes_version) do
    '1.6.6'
  end

  let(:pre_condition) do
    "
      class kubernetes{
        $_authorization_mode = ['RBAC']
        $version = '#{kubernetes_version}'
        $cloud_provider = 'aws'
        $cluster_name = 'cluster1'
      }
      define kubernetes::apply(
        $manifests,
      ){}
    "
  end

  let(:manifests) do
    catalogue.resource('Kubernetes::Apply', 'cluster-autoscaler').send(:parameters)[:manifests]
  end

  context 'with defaults' do
    it 'is valid yaml' do
      manifests.each do |manifest|
        YAML.parse manifest
      end
    end

    it 'has image set' do
      expect(manifests[0]).to match(%{^[-\s].*image: [^:]+:[^:]+$})
    end

    it 'has resources set' do
      expect(manifests[0]).to match(%{^[-\s].*cpu: [0-9]+})
      expect(manifests[0]).to match(%{^[-\s].*memory: [0-9]+})
    end

    it 'has asg configured' do
      expect(manifests[0]).to match(%{3:6:cluster1-kubernetes-worker})
    end

    it 'has cloud_provider configured' do
      expect(manifests[0]).to match(%{cloud-provider=aws})
    end

    it 'has AWS_REGION set' do
      expect(manifests[0]).to match(%r{value: eu-west-1$})
    end

    it 'has host network set' do
      expect(manifests[0]).to match(%r{hostNetwork: true$})
    end

    it 'has master toleration set' do
      expect(manifests[0]).to match(%r{tolerations:\s+- key: "node-role\.kubernetes\.io\/master"\s+operator: "Exists"\s+effect: "NoSchedule"})
    end

    it 'has critical addon toleration set' do
      expect(manifests[0]).to match(%r{- key: "CriticalAddonsOnly"\s+operator: "Exists"})
    end

    it 'has master node affinity set' do
      expect(manifests[0]).to match(%r{nodeAffinity:\s+requiredDuringSchedulingIgnoredDuringExecution:\s+nodeSelectorTerms:\s+- matchExpressions:\s+- key: "node-role\.kubernetes\.io\/master"\s+operator: "Exists"})
    end
  end

  context 'with kubernetes 1.5' do
    let(:kubernetes_version) do
      '1.5.0'
    end
    it 'uses correct image version' do
      expect(manifests[0]).to match(%r{gcr.io/google_containers/cluster-autoscaler:v0.4.0})
    end
  end

  context 'with kubernetes 1.6' do
    let(:kubernetes_version) do
      '1.6.0'
    end
    it 'uses correct image version' do
      expect(manifests[0]).to match(%r{gcr.io/google_containers/cluster-autoscaler:v0.5.4})
    end
  end

  context 'with kubernetes 1.7' do
    let(:kubernetes_version) do
      '1.7.0'
    end
    it 'uses correct image version' do
      expect(manifests[0]).to match(%r{gcr.io/google_containers/cluster-autoscaler:v0.6.0})
    end
  end

  context 'with kubernetes 1.8' do
    let(:kubernetes_version) do
      '1.8.0'
    end
    it 'uses correct image version' do
      expect(manifests[0]).to match(%r{gcr.io/google_containers/cluster-autoscaler:v1.0.0})
    end
  end

  context 'with kubernetes 1.9' do
    let(:kubernetes_version) do
      '1.9.0'
    end
    it 'uses correct image version' do
      expect(manifests[0]).to match(%r{gcr.io/google_containers/cluster-autoscaler:v1.1.0})
    end
  end

  context 'with kubernetes 1.10' do
    let(:kubernetes_version) do
      '1.10.0'
    end
    it 'uses correct image version' do
      expect(manifests[0]).to match(%r{gcr.io/google_containers/cluster-autoscaler:v1.2.0})
    end
  end
end
