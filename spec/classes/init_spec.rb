require 'spec_helper'
describe 'aws_ebs' do
  context 'with default values for all parameters' do
    it { should contain_class('aws_ebs') }
    it {
      should contain_package('curl')
      should contain_package('util-linux')
      should contain_package('xfsprogs')
    }
  end
end
