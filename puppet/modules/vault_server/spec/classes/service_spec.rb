require 'spec_helper'

describe 'vault_server::service' do
    let(:pre_condition) do
        [
            'include vault_server',
            'include vault_server::install'
        ]
    end

    let :systemd_dir do
        '/etc/systemd/system'
    end

    context 'with default values for all parameters' do
        it { should contain_class('vault_server::service') }

        it 'should create vault assets unit' do
            should contain_file(systemd_dir+'/vault-assets.service').with(
                :mode => '0644',
            )
            should contain_exec('vault-assets-systemctl-daemon-reload').with_command(
                'systemctl daemon-reload'
            )
            should contain_service('vault-assets.service').with(
                :ensure => 'stopped',
                :enable => false,
            )
        end

        it 'should create vault unsealer unit' do
            should contain_file(systemd_dir+'/vault-unsealer.service').with(
                :mode => '0644',
            )
            should contain_exec('vault-unsealer-systemctl-daemon-reload').with_command(
                'systemctl daemon-reload'
            )
            should contain_service('vault-unsealer.service').with(
                :ensure => 'running',
                :enable => true,
            )
        end

        it 'should create vault unit' do
            should contain_file(systemd_dir+'/vault.service').with(
                :mode => '0644',
            )
            should contain_exec('vault-systemctl-daemon-reload').with_command(
                'systemctl daemon-reload'
            )
            should contain_service('vault.service').with(
                :ensure => 'running',
                :enable => true,
            )
        end
    end
end
