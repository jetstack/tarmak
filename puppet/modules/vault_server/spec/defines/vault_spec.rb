require 'spec_helper'

describe 'vault_server::vault', :type => :define do

  let(:title) do
    'vault'
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

  context 'should create a vault script' do
    it do
        should contain_file(script_file).with_content(/export PATH=\$PATH:\/usr\/local\/bin/)
    end
  end
end
