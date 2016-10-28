require 'spec_helper'
describe 'etcd' do
  context 'with default values for all parameters' do
    it { should contain_class('etcd') }
  end
end
