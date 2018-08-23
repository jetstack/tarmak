require 'spec_helper'
describe 'site_module' do
  context 'with default values for all parameters' do
    it { should contain_class('site_module') }
  end
end
