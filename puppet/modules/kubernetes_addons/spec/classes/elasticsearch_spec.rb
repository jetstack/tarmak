require 'spec_helper'
describe 'kubernetes_addons::elasticsearch' do
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
    catalogue.resource('Kubernetes::Apply', 'elasticsearch').send(:parameters)[:manifests]
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

    it 'have empty_dir' do
      expect(manifests[1]).to match(%{^[-\s].*emptyDir: \{\}})
    end
  end

  context 'with persistent storage' do
    let(:params) do
      {
        'persistent_storage': true,
      }
    end

    it 'has a pvc claim volume' do
      expect(manifests[1]).to match(%{^[-\s].*claimName: elasticsearch-logging})
    end
    it 'has a pvc object' do
      expect(manifests[1]).to match(%r{kind: PersistentVolumeClaim})
      expect(manifests[1]).to match(%r{volume.beta.kubernetes.io/storage-class: fast})
      expect(manifests[1]).to match(%r{storage: 20Gi})
    end
  end

  context 'with node port' do
    let(:params) do
      {
        'node_port': 32920,
      }
    end

    it 'has a node port svc' do
      expect(manifests[0]).to match(%r{ type: NodePort})
      expect(manifests[0]).to match(%r{ nodePort: 32920})
    end
  end
end
