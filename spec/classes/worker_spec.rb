require 'spec_helper'

describe 'kubernetes::worker' do
  context 'with default values for all parameters' do
    it { should contain_class('kubernetes::kubelet') }
  end
end
