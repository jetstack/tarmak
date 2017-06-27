require 'spec_helper'

describe 'tarmak::master' do
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
