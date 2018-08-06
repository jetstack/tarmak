require 'spec_helper'

describe 'vault_server::assets_service', :type => :define do

  let(:title) do
    'vault-assets'
  end

  let(:service_name) do
    "#{title}.service"
  end

  let(:service_file) do
    "/etc/systemd/system/#{service_name}"
  end

  let(:pre_condition) {[
    "
class{'vault_server':
  token => 'token1'
}
    "
  ]}

  let(:params) do
    {
       :vault_tls_cert_path => '/tmp/asset/crt',
       :vault_tls_key_path => '/tmp/asset/pem',
       :vault_tls_ca_path => '/tmp/asset/ca',
    }
  end

  context 'should create a vault asset service' do
    it do
      #should contain_service(service_name)
      #should contain_file(service_file).with_content(%r{aws s3 cp /tmp/asset/crt /etc/vault/tls/tls.pem && aws s3 cp /tmp/asset/pem /etc/vault/tls/tls-key.pem && chmod 0600 /etc/vault/tls/tls-key.pem && aws s3 cp /tmp/asset/ca /etc/vault/tls/ca.pem})
    end
  end
end
