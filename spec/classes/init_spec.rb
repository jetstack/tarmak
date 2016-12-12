require 'spec_helper'

describe 'vault_client' do
  let(:pre_condition) {[
    """class systemd{
    exec {'systemctl-daemon-reload':}
}
    """
  ]}

  context 'with none of init_token and token specified' do
    it do
      is_expected.to compile.and_raise_error(/provide at least one of/)
    end
  end

  context 'with init_token and token specified' do
    let(:params) { {:init_token => 'ab', :token => 'cd'} }
    it do
      is_expected.to compile.and_raise_error(/must provide either \$init_token or \$token/)
    end
  end

  context "with token 'test-token'" do
    let(:params) { {:token => 'test-token'} }
    it do
      is_expected.to contain_file('/etc/vault/token').with(
        'mode' => '0600',
        'content' => 'test-token'
      )
    end
    it do
      is_expected.to contain_file('/etc/vault/config').with_content(/#{Regexp.escape('VAULT_ADDR=http://127.0.0.1:8200')}/)
    end
    it do
      is_expected.not_to contain_file('/etc/vault/config').with_content(/#{Regexp.escape('VAULT_INIT_ROLE=')}/)
    end
    it do
      is_expected.not_to contain_file('/etc/vault/config').with_content(/#{Regexp.escape('VAULT_INIT_POLICIES=')}/)
    end
  end

  context "with init_token 'test-init-token'" do
    let(:params) do {
       :init_token => 'test-init-token',
       :init_role => 'test-role',
       :init_policies => ['test-policy1', 'test-policy2'],
      }
    end
    it do
      is_expected.to contain_file('/etc/vault/init-token').with(
        'mode' => '0600',
        'content' => 'test-init-token'
      )
    end
    it do
      is_expected.to contain_file('/etc/vault/config').with_content(/#{Regexp.escape('VAULT_ADDR=http://127.0.0.1:8200')}/)
    end
    it do
      is_expected.to contain_file('/etc/vault/config').with_content(/#{Regexp.escape('VAULT_INIT_ROLE=test-role')}/)
    end
    it do
      is_expected.to contain_file('/etc/vault/config').with_content(/#{Regexp.escape('VAULT_INIT_POLICIES=test-policy1,test-policy2')}/)
    end
  end

  context 'with custom version 1.2.3' do
    let(:params) { {:version => '1.2.3', :token => 'ab'} }

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
