require 'spec_helper'

describe 'kubernetes::dns' do
  let(:pre_condition) do
    [
      'include kubernetes::apiserver'
    ]
  end

  context 'with default values for all parameters' do
    it { should contain_class('kubernetes::dns') }
  end
end
