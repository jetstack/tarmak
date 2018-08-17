require 'spec_helper'

describe 'vault_server::install' do
    let(:pre_condition) do
        [
            'include vault_server'
        ]
    end

    let :version do
        '0.9.5'
    end

    context 'with default values for all parameters' do
        it { should contain_class('vault_server::install') }

        it 'should download and install the vault server' do
            should contain_file('/opt/vault-'+version).with(
                :ensure => 'directory',
            )
            should contain_file('/opt/vault-'+version+'/vault')
            should contain_file('/usr/local/bin/vault').with(
                :ensure => 'link',
            )
        end

        it 'should attempt to download vault unsealer' do
            should contain_file('/usr/local/bin/download-vault-unsealer.sh').with(
                :mode => '0755',
            )
            should contain_exec('download-vault-unsealer-script-run').with_command(
                '/usr/local/bin/download-vault-unsealer.sh'
            )
        end

        it 'should create vault profile script' do
            should contain_file('/etc/profile.d/vault.sh').with(
                :mode => '0644',
            )
        end
    end
end
