require 'spec_helper'

describe 'kubernetes::master' do
  context 'with default values for all parameters' do
    it { should contain_class('kubernetes::apiserver') }
    it { should contain_class('kubernetes::scheduler') }
    it { should contain_class('kubernetes::controller_manager') }
    it { should contain_class('kubernetes::kubelet') }
    it { should contain_class('kubernetes::proxy') }
    it { should contain_class('kubernetes::storage_classes') }
    it { should contain_class('kubernetes::dns') }
  end

  context 'with disabled kubelet' do
    let(:params) { {'disable_kubelet' => true } }
    it { should_not contain_class('kubernetes::kubelet') }
  end

  context 'with disabled kube-proxy' do
    let(:params) { {'disable_proxy' => true } }
    it { should_not contain_class('kubernetes::proxy') }
  end
end
