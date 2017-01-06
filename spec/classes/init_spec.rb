require 'spec_helper'
describe 'calico' do
  context 'with a supplied etcd cluster array with one node' do
  let(:params) {
    {
      :etcd_cluster => ['etcd1']
    }
  }
    it do
      should contain_class('calico')
      should contain_file('/usr/local/sbin/sourcedestcheck.sh').with({
        'ensure' => 'file',
        'mode'   => '0755',
      })
    end
  end

  context 'with a supplied etcd cluster array with five nodes, and aws set to false' do
  let(:params) {
    {
      :etcd_cluster => ['etcd1','etcd2','etcd3','etcd4','etcd5'],
      :aws          => false
    }
  }
    it do
      should contain_class('calico')
      should_not contain_file('/usr/local/sbin/sourcedestcheck.sh')
    end
  end
end
