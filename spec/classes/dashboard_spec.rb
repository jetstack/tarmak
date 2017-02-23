require 'spec_helper'
describe 'kubernetes_addons::dashboard' do
  let(:pre_condition) do
    "
      class kubernetes{}
      define kubernetes::apply(
        $manifests,
      ){}
    "
  end

  let(:manifests) do
    catalogue.resource('Kubernetes::Apply', 'kube-dashboard').send(:parameters)[:manifests]
  end

  context 'with defaults' do
    it 'be valid yaml' do
      manifests.each do |manifest|
        YAML.parse manifest
      end
    end

    it 'have image set' do
      expect(manifests[1]).to match(%{^[-\s].*image: [^:]+:[^:]+$})
    end

    it 'have resources set' do
      expect(manifests[1]).to match(%{^[-\s].*cpu: [0-9]+})
      expect(manifests[1]).to match(%{^[-\s].*memory: [0-9]+})
    end
  end
end
