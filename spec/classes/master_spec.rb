require 'spec_helper'

describe 'kubernetes::master' do
  context 'with default values for all parameters' do
    it { should contain_class('kubernetes::apiserver') }
    it { should contain_class('kubernetes::scheduler') }
    it { should contain_class('kubernetes::controller_manager') }
  end
end
