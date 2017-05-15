require 'spec_helper'

describe 'kubernetes::apply', :type => :define do

  let(:title) do
    'test1'
  end

  let(:pre_condition) {[
    'include kubernetes::apiserver',
  ]}

  let(:service_name) do
    "kubectl-apply-#{title}.service"
  end

  context 'not running on kubernetes master' do
    let(:pre_condition) {[]}
    it { should compile.and_raise_error(/only be used on the kubernetes master/) }
  end

  context 'running on kubernetes master' do
    context 'type == manifests' do
      let :params do
        {
          :type => 'manifests',
        }
      end
      it do
        should contain_service(service_name)
        should contain_file("/etc/systemd/system/#{service_name}")
        should contain_file("/etc/kubernetes/apply/#{title}.yaml")
        should_not contain_concat("/etc/kubernetes/apply/#{title}.yaml")
      end
    end
    context 'type == concat' do
      let :params do
        {
          :type => 'concat',
        }
      end
      it do
        should contain_service(service_name)
        should contain_file("/etc/systemd/system/#{service_name}")
        should contain_concat("/etc/kubernetes/apply/#{title}.yaml")
        should_not contain_file("/etc/kubernetes/apply/#{title}.yaml")
      end
    end
  end

  context ' on kubernetes master' do
    let(:pre_condition) {[
      '
      include kubernetes::apiserver
      '
    ]}

    it do
      should contain_service(service_name)
      should contain_file("/etc/systemd/system/#{service_name}")
      should contain_file("/etc/kubernetes/apply/#{title}.yaml")
    end
  end
end
