require 'spec_helper'
describe 'vault_client' do
  context 'with default values for all parameters' do
    it do
      should contain_class('vault_client')
    end
  end
  context 'with custom version 1.2.3' do
    let(:params) { {:version => '1.2.3'} }

    it do
      is_expected.to contain_archive('/tmp/vault.zip').with(
        'source' => 'https://releases.hashicorp.com/vault/1.2.3/vault_1.2.3_linux_amd64.zip',
      )
    end

    it do
      is_expected.to contain_file('/opt/vault-1.2.3/vault').with(
        'mode' => '0755',
      )
    end
  end
end
