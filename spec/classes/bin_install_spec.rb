require 'spec_helper'
describe 'calico::bin_install' do
  context 'with defaults' do
    let(:pre_condition) { "class calico { $etcd_cluster = ['etcd1'] }" }

    it do
      should contain_class('calico::bin_install')
      should contain_file('/opt/cni/bin/calico').with('mode' => '0755',)
      should contain_calico__wget_file('calico').with(
        'destination_dir' => '/opt/cni/bin',
      )
      should contain_calico__wget_file('calico-ipam').with(
        'destination_dir' => '/opt/cni/bin'
      )
    end
  end
  context 'with custom version' do
    let(:pre_condition) { "class calico { $etcd_cluster = ['etcd1'] }" }
    let(:params) {
      {
        :bin_version => 'v5.6.7'
      }
    }

    it do
      should contain_class('calico::bin_install')
      should contain_file('/opt/cni/bin/calico').with('mode' => '0755',)
      should contain_calico__wget_file('calico').with(
        'destination_dir' => '/opt/cni/bin',
        'url'             => 'https://github.com/projectcalico/cni-plugin/releases/download/v5.6.7/calico'
      )
      should contain_calico__wget_file('calico-ipam').with(
        'destination_dir' => '/opt/cni/bin',
        'url'             => 'https://github.com/projectcalico/cni-plugin/releases/download/v5.6.7/calico-ipam'
      )
    end
  end
end
