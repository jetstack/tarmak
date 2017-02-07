require 'spec_helper'

describe 'kubernetes::master' do
  context 'with default values for all parameters' do
    it { should contain_class('kubernetes::apiserver') }
    it { should contain_class('kubernetes::scheduler') }
    it { should contain_class('kubernetes::controller_manager') }
    it { should contain_class('kubernetes::kubelet') }
  end

  context 'with disabled kubelet' do
    let(:params) { {'disable_kubelet' => true } }
    it { should_not contain_class('kubernetes::kubelet') }
  end
end
