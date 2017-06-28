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
      expect(manifests[0]).to match(%{3:6:kubernetes-cluster1-worker})
    end

    it 'has cloud_provider configured' do
      expect(manifests[0]).to match(%{cloud-provider=aws})
    end

    it 'has cert path set' do
      expect(manifests[0]).to match(%{path: /etc/ssl/certs})
    end

    it 'has AWS_REGION set' do
      expect(manifests[0]).to match(%r{value: eu-west-1$})
    end
  end

  context 'with kubernetes 1.5' do
    let(:kubernetes_version) do
      '1.5.6'
    end
    it 'uses correct image version' do
      expect(manifests[0]).to match(%r{gcr.io/google_containers/cluster-autoscaler:v0.4.0})
    end
  end

  context 'with kubernetes 1.6' do
    let(:kubernetes_version) do
      '1.6.6'
    end
    it 'uses correct image version' do
      expect(manifests[0]).to match(%r{gcr.io/google_containers/cluster-autoscaler:v0.5.4})
    end
  end

  context 'with kubernetes 1.7' do
    let(:kubernetes_version) do
      '1.7.1'
    end
    it 'uses correct image version' do
      expect(manifests[0]).to match(%r{gcr.io/google_containers/cluster-autoscaler:v0.6.0})
    end
  end
end
