require 'spec_helper'

describe 'role: master' do
  let(:facts) do
    {
      :tarmak_role          => 'master',
      :tarmak_type_instance => 'tarmak',
    }
  end

  it 'sets up docker' do
    is_expected.to contain_package('docker')
    is_expected.to contain_class('site_module::docker')
    is_expected.to contain_class('site_module::docker_config')
  end

  it 'sets up master tarmak' do
    is_expected.to contain_class('tarmak::master')
  end

  it 'sets up vault_client' do
    is_expected.to contain_class('vault_client').with_init_token('init-token1')
    is_expected.to contain_class('vault_client').with_init_role('cluster1-master')
    is_expected.to contain_class('vault_client').with_server_url('https://vault.domain-zone.root:8200')
    is_expected.to contain_class('vault_client').with_ca_cert_path('/etc/vault/ca.pem')
  end
end
