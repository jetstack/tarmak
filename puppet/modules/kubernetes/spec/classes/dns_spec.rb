require 'spec_helper'

describe 'kubernetes::dns' do
  context 'with default values for all parameters' do
    let(:pre_condition) do
      "
        class{'kubernetes': version => '1.9.0'}
        define kubernetes::apply(
          $manifests,
        ){
          kubernetes::addon_manager_labels($manifests[0])
        }
        define kubernetes::delete(){}
      "
    end

    let(:manifests) do
      catalogue.resource('Kubernetes::Apply', 'kube-dns').send(:parameters)[:manifests]
    end

    it { should contain_class('kubernetes::dns') }

    it 'be valid yaml' do
      manifests.each do |manifest|
        YAML.parse manifest
      end
    end

    it 'should write manifests' do
      expect(manifests.join('\n---\n')).to match(%{--domain=cluster\.local\.})
      expect(manifests.join('\n---\n')).to match(%{clusterIP: 10\.254\.0\.10})
    end
  end

  context 'with version 1.11' do
    let(:pre_condition) do
      "
        class{'kubernetes': version => '1.11.0'}
        define kubernetes::apply(
          $manifests,
        ){
          kubernetes::addon_manager_labels($manifests[0])
        }
        define kubernetes::delete(){}
      "
    end

    let(:manifests) do
      catalogue.resource('Kubernetes::Apply', 'coredns').send(:parameters)[:manifests]
    end

    it 'be valid yaml' do
      manifests.each do |manifest|
        YAML.parse manifest
      end
    end

    it { should contain_class('kubernetes::dns') }

  end
end
