require 'spec_helper'

describe 'consul::config' do
    let(:pre_condition) do
        [
            'include consul'
        ]
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
                .with_content(/#{Regexp.escape('"acl_master_token" : "${consul_master_token}",')}/)
        end

        it 'should install vault.hcl file' do
            should contain_file('/etc/vault/vault.hcl').with(
                :mode => '0600',
            )
            should contain_file('/etc/vault/vault.hcl')
                .with_content(/#{Regexp.escape('path = "vault-${environment}/"')}/)
        end

        it 'should install consul master token' do
            should contain_file('/etc/consul/master-token').with(
                :mode => '0600',
            )
            should contain_file('/etc/consul/master-token')
                .with_content(/#{Regexp.escape('CONSUL_HTTP_TOKEN=${consul_master_token}')}/)
        end
    end
end
