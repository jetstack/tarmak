require 'spec_helper'

describe 'consul' do
  it do
    should contain_class('consul::install')
  end
end
