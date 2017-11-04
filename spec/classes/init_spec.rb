require 'spec_helper'
describe 'calico' do
  let(:pre_condition) do
    "
      class kubernetes{
        $_authorization_mode = ['RBAC']
        $version = '1.7.10'
      }
      define kubernetes::apply(
        $manifests,
      ){}
    "
  end

  let(:manifests) do
    catalogue.resource('Kubernetes::Apply', 'disable-srcdest-node').send(:parameters)[:manifests]
  end

  context 'on cloud_provider aws' do
    let(:params) {
      {
        :cloud_provider => 'aws'
      }
    }

    it 'contain source_destination check' do
      should contain_class('calico')
      should contain_class('calico::disable_source_destination_check')
    end

    it 'removes old systemd unit' do
      should contain_file('/opt/bin/disable-source-destination-check.sh').with({'ensure' => 'absent'})
      should contain_file('/etc/systemd/system/disable-source-destination-check.service').with({'ensure' => 'absent'})
    end

    context 'on masters' do
      let(:facts) do
        {
          :ec2_metadata => {
            'placement' => {
              'availability-zone' => 'eu-west-1a',
            }
          }
        }
      end

      let(:pre_condition) do
        "
      class kubernetes{
        $_authorization_mode = ['RBAC']
        $version = '1.7.10'
      }
      class kubernetes::apiserver{}
      include kubernetes::apiserver
      define kubernetes::apply(
        $manifests,
      ){}
        "
      end

      it 'is valid yaml' do
        manifests.each do |manifest|
          YAML.parse manifest
        end
      end

      it 'has deployment of k8s-srcdest' do
        expect(manifests[0]).to match(%{ image: ottoyiu/k8s-ec2-srcdst:v})
        expect(manifests[0]).to match(%{value: "eu-west-1"})
      end
    end
  end

  context 'on other cloud_providers' do
    let(:params) {
      {
        :cloud_provider => ''
      }
    }

    it do
      should contain_class('calico')
      should_not contain_class('calico::disable_source_destination_check')
    end
  end
end
