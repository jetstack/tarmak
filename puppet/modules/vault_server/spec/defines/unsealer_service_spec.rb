require 'spec_helper'

describe 'vault_server::unsealer_service', :type => :define do

  let(:title) do
    'vault-unsealer'
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
       :region                        => 'region1',
       :vault_unsealer_kms_key_id     => 'key_id1',
       :vault_unsealer_ssm_key_prefix => 'key_prefix1',
    }
  end

  context 'should create a vault unsealer service' do
    it do
      #should contain_service(service_name)
      #should contain_file(service_file).with_content(/REGION=region1/)
      #should contain_file(service_file).with_content(/Environment=VAULT_UNSEALER_AWS_KMS_KEY_ID=key_id1/)
      #should contain_file(service_file).with_content(/Environment=VAULT_UNSEALER_AWS_SSM_KEY_PREFIX=key_prefix1/)
    end
  end
end
