require 'spec_helper'
describe 'fluent_bit' do
  context 'with default values for all parameters' do
    it { should contain_class('fluent_bit') }
  end
  
  context 'on cloud_provider aws' do
    let(:params) {
      {
        :cloud_provider => 'aws'
      }
    }
  end
end
