require 'spec_helper'

describe 'vault_client::vault_service', :type => :define do

  let(:title) do
    'vault'
  end

  let(:service_name) do
    "#{title}.service"
  end

  let(:service_file) do
    "/etc/systemd/system/#{service_name}"
  end

  let(:pre_condition) {[
    "
class{'vault_client':
  token => 'token1'
}
    "
  ]}

  let(:params) do
    {
       :region => 'region1',
    }
  end

  context 'should create a vault service' do
    it do
      should contain_service(service_name)
      should contain_file(service_file).with_content(/AWS_REGION=region1/)
    end
  end
end
