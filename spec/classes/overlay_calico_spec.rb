require 'spec_helper'

describe 'puppernetes::overlay_calico' do
  let(:pre_condition) {[
    """
class{'vault_client': token => 'test-token'}
class{'puppernetes': role => 'master'}
include kubernetes::master
"""
  ]}

  context 'without params' do
    it do
      should contain_calico__ip_pool('10.234.0.0/16')
    end
  end
end
