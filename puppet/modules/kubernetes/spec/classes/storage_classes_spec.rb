require 'spec_helper'

describe 'kubernetes::storage_classes' do
  let(:pre_condition) do
    [
      "class{'kubernetes':
         version => '#{version}',
         cloud_provider => '#{cloud_provider}',
      }",
      'include kubernetes::apiserver'
    ]
  end
end
