require 'spec_helper'
describe 'fluent_bit' do

  let(:pre_condition) do
    [
      'include kubernetes::apiserver'
    ]
  end
  let(:pre_condition) do
    """
      class{'fluent_bit': ensure => 'present'}
    """
  end

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
