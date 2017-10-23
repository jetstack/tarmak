require 'spec_helper'

describe 'tarmak::overlay_calico' do
  let(:pre_condition) {[
    """
class{'vault_client': token => 'test-token'}
class{'tarmak': role => 'master'}
include kubernetes::master
"""
  ]}

  context 'without params' do
    it do
      should contain_class('calico')
      should contain_class('calico::config')
      should contain_class('calico::node')
      should contain_class('calico::policy_controller')
    end
  end
end
