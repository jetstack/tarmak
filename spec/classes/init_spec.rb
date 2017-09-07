require 'spec_helper'

describe 'vault_client' do
    version = "0.8.6"
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
        end
    end

    context "with init_token 'test-init-token'" do
        let(:params) do {
            :init_token => 'test-init-token',
            :init_role => 'test-master',
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
        end
    end

    context "with custom version #{version}" do
        let(:params) { {:version => version, :token => 'ab'} }

        it do
            is_expected.to contain_file('/opt/bin/vault-helper').with(
                'target' => "/opt/vault-helper-#{version}/vault-helper"
            )
        end

        it do
            is_expected.to contain_file('/opt/bin/vault-helper').with(
                'mode' => '0755',
            )
        end
    end
end
