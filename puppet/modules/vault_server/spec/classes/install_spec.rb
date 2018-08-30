require 'spec_helper'

describe 'vault_server::install' do
  let(:pre_condition) do
    """
      class{'vault_server':
        cloud_provider => 'aws',
        consul_master_token => 'master_token',
        environment => 'env',
      }
    """
  end

  context 'with default values for all parameters' do
    it { should contain_class('vault_server::install') }

    it 'should download and install the vault server' do
      should contain_file('/opt/vault-0.9.5').with(
        :ensure => 'directory',
      )
      should contain_file('/opt/vault-0.9.5/vault')
      should contain_file('/opt/bin/vault').with(
        :ensure => 'link',
        :target => '/opt/vault-0.9.5/vault',
      )
    end

    it 'should attempt to download vault unsealer' do
      should contain_file('/opt/vault-0.9.5/vault-unsealer').with(
        :mode => '0755',
      )
      should contain_file('/opt/bin/vault-unsealer').with(
        :ensure => 'link',
        :target => '/opt/vault-0.9.5/vault-unsealer',
      )
    end

    it 'should create vault profile script' do
      should contain_file('/etc/profile.d/vault.sh').with(
        :mode => '0644',
      )
    end

    it 'should install vault.hcl file' do
      should contain_file('/etc/vault/vault.hcl').with(
        :mode => '0600',
      )
      should contain_file('/etc/vault/vault.hcl')
        .with_content(/#{Regexp.escape('path = "vault-env/"')}/)
    end
  end
end
