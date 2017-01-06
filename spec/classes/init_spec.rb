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
    it do
      should contain_class('calico::bin_install')
      should contain_class('calico::lo_install')
      should contain_class('calico::config')
    end
    it do
      should contain_class('calico::node').with({
        'etcd_endpoints' => 'http://etcd1:2359',
      })
    end
  end

  context 'with a supplied etcd cluster array with five nodes, tls, custom port, and aws set to false' do
  let(:params) {
    {
      :etcd_cluster         => ['etcd1','etcd2','etcd3','etcd4','etcd5'],
      :aws                  => false,
      :tls                  => true,
      :etcd_overlay_port    => 5678
    }
  }
    it do
      should contain_class('calico')
    end
    it do
      should contain_class('calico::node').with({
        'etcd_endpoints' => 'https://etcd1:5678,https://etcd2:5678,https://etcd3:5678,https://etcd4:5678,https://etcd5:5678'
      })
    end
    it do
      should_not contain_file('/usr/local/sbin/sourcedestcheck.sh')
    end
  end
end
