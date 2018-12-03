require 'spec_helper'

describe 'role: worker' do
  let(:facts) do
    {
      :tarmak_role          => 'worker',
    }
  end

  it 'sets up docker' do
    is_expected.to contain_package('docker')
    is_expected.to contain_class('site_module::docker')
    is_expected.to contain_class('site_module::docker_config')
    is_expected.to contain_class('site_module::docker_storage')
  end

  it 'sets up docker storage before everything else' do
    is_expected.to contain_file('/etc/sysconfig/docker-storage-setup').that_comes_before('Service[docker.service]')
  end

  it 'sets up kubelet after docker service' do
    is_expected.to contain_service('docker.service').that_comes_before('Service[kubelet.service]')
  end

  it 'sets up worker tarmak' do
    is_expected.to contain_class('tarmak::worker')
  end

  it 'sets up vault_client' do
    is_expected.to contain_class('vault_client').with_init_token('init-token1')
    is_expected.to contain_class('vault_client').with_init_role('cluster1-worker')
    is_expected.to contain_class('vault_client').with_server_url('https://vault.domain-zone.root:8200')
    is_expected.to contain_class('vault_client').with_ca_cert_path('/etc/vault/ca.pem')
  end
end
