require 'spec_helper'

describe 'puppernetes::master' do
  let(:facts) do
    @default_facts
  end

  let(:pre_condition) {[
    """
class{'vault_client': token => 'test-token'}
"""
  ]}

  context 'without params' do
    it do
      is_expected.to compile
    end
  end
end
