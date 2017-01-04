require 'spec_helper'
describe 'calico' do
  context 'with a supplied etcd cluster array' do
  let(:facts) { 
    {
      :operatingsystem => 'CentOS',
      :kernel => 'Linux'
    } 
  }
  let(:params) {
    {
      :etcd_cluster => ['etcd1']
    }
  }
    it { should contain_class('calico') }
  end
end
