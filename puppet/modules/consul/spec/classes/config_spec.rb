require 'spec_helper'

describe 'consul::config' do
  let(:pre_condition) do
    """
      class{'consul':
        cloud_provider => 'aws',
        consul_master_token => 'master_token',
        environment => 'env',
      }
    """
  end

  context 'with default values for all parameters' do
    it { should contain_class('consul::config') }

    it 'should install consul config' do
      should contain_file('/etc/consul').with(
        :ensure => 'directory',
      )
      should contain_file('/etc/consul/consul.json').with(
        :mode => '0600',
      )
      should contain_file('/etc/consul/consul.json')
        .with_content(/"acl_master_token" : "master_token"/)
      should contain_file('/etc/consul/consul.json')
        .with_content(/[provider=aws tag_key=VaultCluster tag_value=env}]/)
    end

    it 'should install consul master token' do
      should contain_file('/etc/consul/master-token').with(
        :mode => '0600',
      )
      should contain_file('/etc/consul/master-token')
        .with_content(/#{Regexp.escape('CONSUL_HTTP_TOKEN=master_token')}/)
    end
  end
end
