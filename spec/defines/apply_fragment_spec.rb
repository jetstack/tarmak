require 'spec_helper'

describe 'kubernetes::apply_fragment', :type => :define do
  let(:pre_condition) {[
    'include kubernetes::apiserver',
  ]}

  let(:title) do
    'test1'
  end

  let :params do
    {
      :manifests => ["asdf"],
      :order     => "1",
    }
  end

  let(:service_name) do
    "kubectl-apply-#{title}.service"
  end

  context 'not running on kubernetes master' do
    let(:pre_condition) {[]}
    it { should compile.and_raise_error(/only be used on the kubernetes master/) }
  end

  context 'running on kubernetes master' do
    context 'type == concat' do
      it do
        should contain_service(service_name)
        should contain_concat__fragment("/etc/systemd/system/#{service_name}")
	    .with_content(/^Description=kubectl apply #{service_name}$/)
        should contain_file("/etc/kubernetes/apply/#{title}.yaml")
      end
    end
  end
  context 'running on kubernetes master' do
    context 'type == unknown' do
      it { should compile.and_raise_error(/Unknown type parameter: 'unknown'/) }
    end
  end
end
