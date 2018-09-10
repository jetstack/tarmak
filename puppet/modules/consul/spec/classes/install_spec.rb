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
      should contain_file('/opt/consul-exporter-0.3.0').with(
        :ensure => 'directory',
      )
      should contain_file('/opt/consul-exporter-0.3.0/consul_exporter').with(
        :ensure => 'file',
        :mode => '0755',
      )
      should contain_file('/opt/bin/consul_exporter').with(
        :ensure => 'link',
        :target => '/opt/consul-exporter-0.3.0/consul_exporter',
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

    it 'should install consul backinator' do
      should contain_file('/opt/consul-backinator-1.3').with(
        :ensure => 'directory',
      )
      should contain_file('/opt/consul-backinator-1.3/consul-backinator').with(
        :ensure => 'file',
        :mode => '0755',
      )
      should contain_file('/opt/bin/consul-backinator').with(
        :ensure => 'link',
        :target => '/opt/consul-backinator-1.3/consul-backinator',
      )
    end
  end
end
