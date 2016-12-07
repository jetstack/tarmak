require 'spec_helper_acceptance'

describe 'vault_client class' do

  context 'default parameters' do
    # Using puppet_apply as a helper
    it 'should work with no errors based on the example' do
      pp = <<-EOS
        class {'vault_client':
          version => '0.6.2',
        }
      EOS

      # Run it twice and test for idempotency
      apply_manifest(pp, :catch_failures => true)
      expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero
    end

    it do
      show_result = shell('vault version')
      expect(show_result.stdout).to match(/Vault v0\.6\.2/)
    end
  end
end
