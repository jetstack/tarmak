require 'spec_helper'
describe 'calico::lo_install' do
  context 'with defaults' do
    it do
      should contain_class('calico::lo_install')
      should contain_archive('download and extract cni-lo').with(
        'creates' => '/opt/cni/bin/loopback'
      )
    end
  end

  context 'with custom version' do
    let(:params) {
      {
        :cni_version => 'v0.6.7'
      }
    }

    it do
      should contain_class('calico::lo_install')
      should contain_archive('download and extract cni-lo').with(
        'creates' => '/opt/cni/bin/loopback',
        'source'  => 'https://github.com/containernetworking/cni/releases/download/v0.6.7/cni-v0.6.7.tgz'
      )
    end
  end
end
