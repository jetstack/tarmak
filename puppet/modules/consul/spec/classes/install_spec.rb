require 'spec_helper'

describe 'consul::install' do
  let(:pre_condition) do
    [
      'include consul'
    ]
  end

  let :version do
    '1.2.1'
  end

  context 'with default values for all parameters' do
    it { should contain_class('consul::install') }

    it 'should install consul' do
      should contain_file('/opt/consul-'+version).with(
        :ensure => 'directory',
      )
      should contain_file('/opt/consul-'+version+'/consul').with(
        :mode => '0755',
      )
      should contain_file('/opt/bin/consul').with(
        :ensure => 'link',
        :target => '/opt/consul-1.2.1/consul',
      )
    end

    it 'should install consul exporter' do
      should contain_file('/opt/consul_exporter-0.3.0').with(
        :ensure => 'directory',
      )
    end

    it 'should install consul backup script' do
      should contain_file('/opt/consul-1.2.1/consul-backup.sh').with(
        :ensure => 'file',
        :mode => '0755',
      )
      should contain_file('/opt/bin/consul-backup.sh').with(
        :ensure => 'link',
        :target => '/opt/consul-1.2.1/consul-backup.sh',
      )
    end
  end
end
