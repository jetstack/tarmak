require 'spec_helper'
describe 'vault_client' do
  context 'with default values for all parameters' do
    it { should contain_class('vault_client') }
  end
end
