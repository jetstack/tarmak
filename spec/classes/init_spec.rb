require 'spec_helper'


describe 'etcd' do
  let(:pre_condition) {[
    'class vault_client{}'
  ]}
  context 'with default values for all parameters' do
    it { should contain_class('etcd') }
  end
end
