require 'spec_helper'
describe 'aws_es_proxy' do
  context 'with default values for all parameters' do
    it { should contain_class('aws_es_proxy') }
  end
  
  context 'on cloud_provider aws' do
    let(:params) {
      {
        :cloud_provider => 'aws'
      }
    }
  end
end
