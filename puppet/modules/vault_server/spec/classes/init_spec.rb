require 'spec_helper'

describe 'vault_server' do
  let :config_dir do
    '/etc/vault'
  end

  let :lib_dir do
    '/var/lib/vault'
  end

  let :app_name do
    'vault'
  end

  context 'with default values for all parameters' do
    it { should contain_class('vault_server') }

    it 'should create directorys' do
      should contain_file(config_dir).with(
        :ensure => 'directory',
      )
      should contain_file(lib_dir).with(
        :ensure => 'directory',
      )
    end

    it 'should create vault user' do
      should contain_user(app_name).with(
        :home => lib_dir,
      )
    end

  end
end
