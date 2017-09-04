require 'spec_helper'

describe 'vault_client::secret_service', :type => :define do

  let(:title) do
    'test1'
  end

  let(:service_name) do
    "#{title}-secret.service"
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
      :dest_path => '/tmp/dest_path1',
      :secret_path => '/my/secret1',
      :field => 'field1',
      :user => 'user1',
      :group => 'group1',
    }
  end

  context 'should create a vault secert service' do
    it do
      should contain_service(service_name)
      should contain_file(service_file).with_content(/Environment=VAULT_CERT_OWNER=user1:group1/)
      should contain_file(service_file).with_content(%r{/opt/bin/vault-helper read /my/secret1 -f field1 -d /tmp/dest_path1})
    end
  end
end
