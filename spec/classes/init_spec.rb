require 'spec_helper'

describe 'puppernetes' do
  let(:facts) do 
      @default_facts
  end
  context 'without params' do
    it do
      is_expected.to compile
    end
  end
end
