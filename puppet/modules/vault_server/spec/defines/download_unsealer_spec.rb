require 'spec_helper'

describe 'vault_server::download_unsealer', :type => :define do

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
class{'vault_server':
  token => 'token1'
}
    "
  ]}

  context 'should create a vault unsealer download script' do
    it do
        #should contain_file(script_file).with_content(/VAULT_UNSEALER_HASH=7a01a119429b93edecb712aa897f2b22ba0575b7db5f810d4a9a40d993dad1aa/)
    end
  end
end
