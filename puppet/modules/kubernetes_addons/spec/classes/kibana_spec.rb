require 'spec_helper'
describe 'kubernetes_addons::kibana' do
  let(:pre_condition) do
    "
      class kubernetes{}
      define kubernetes::apply(
        Enum['present', 'absent'] $ensure = 'present',
        $manifests,
      ){
        if $manifests and $ensure == 'present' {
          kubernetes::addon_manager_labels($manifests[0])
        }
      }
    "
  end

  let(:manifests) do
    catalogue.resource('Kubernetes::Apply', 'kibana').send(:parameters)[:manifests]
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
