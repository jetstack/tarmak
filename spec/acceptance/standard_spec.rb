require 'spec_helper_acceptance'

describe 'vault_client class' do

  context 'default parameters' do
    # Using puppet_apply as a helper
    it 'should work with no errors based on the example' do
      pp = <<-EOS
        file { '/opt/vault_client/':
          ensure => 'directory',
          owner  => 'vault_client',
          group  => 'root',
        } ->
        class { 'vault_client':
          config_hash => {
              'datacenter' => 'east-aws',
              'data_dir'   => '/opt/vault_client',
              'log_level'  => 'INFO',
              'node_name'  => 'foobar',
              'server'     => true
          }
        }
      EOS

      # Run it twice and test for idempotency
      expect(apply_manifest(pp).exit_code).to_not eq(1)
      expect(apply_manifest(pp).exit_code).to eq(0)
    end

    describe service('vault_client') do
      it { should be_enabled }
    end

    describe command('vault_client version') do
      it { should return_stdout /Consul v0\.2\.0/ }
    end

  end
end
