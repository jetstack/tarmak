require 'spec_helper'

describe 'vault_client::download_unsealer', :type => :define do

  let(:title) do
    'download-vault-unsealer'
  end

  let(:script_name) do
    "#{title}.sh"
  end

  let(:script_file) do
    "/usr/local/bin/#{script_name}"
  end

  let(:pre_condition) {[
    "
class{'vault_client':
  token => 'token1'
}
    "
  ]}

  context 'should create a vault unsealer download script' do
    it do
        should contain_file(script_file).with_content(/curl -sL https:\/\/github.com\/jetstack\/vault-unsealer\/releases\/download\/\$\${VAULT_UNSEALER_VERSION}\/vault-unsealer_\$\${VAULT_UNSEALER_VERSION}_linux_amd64 > \$\${DEST_DIR}\/vault-unsealer/)
    end
  end
end
