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
      :content => "world",
      :order   => "2",
      :target  => "test1",
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
    it do
      should contain_concat__fragment("kubectl-apply-test1")
        .with_content(/^world$/)
    end
  end
end
