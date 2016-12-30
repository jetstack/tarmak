require 'spec_helper'

describe 'puppernetes::etcd' do
  let(:facts) do 
    {
      :path => '/bin:/sbin:/usr/bin:/usr/sbin:/opt/bin'
    }
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
