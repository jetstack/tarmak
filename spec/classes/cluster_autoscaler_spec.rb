require 'spec_helper'
describe 'kubernetes_addons::cluster_autoscaler' do
  let(:pre_condition) do
    "
      class kubernetes{
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
  end
end
