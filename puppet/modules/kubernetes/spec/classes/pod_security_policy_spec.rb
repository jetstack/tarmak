require 'spec_helper'

describe 'kubernetes::pod_security_policy' do
  context 'with default values for all parameters' do
    let(:pre_condition) do
      "
        class{'kubernetes': version => '1.9.0'}
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
      catalogue.resource('Kubernetes::Apply', 'puppernetes-rbac-psp').send(:parameters)[:manifests]
    end

    it { should contain_class('kubernetes::pod_security_policy') }

    it 'be valid yaml' do
      manifests.each do |manifest|
        YAML.parse manifest
      end
    end
  end
end
