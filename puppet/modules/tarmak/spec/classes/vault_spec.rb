require 'spec_helper'

describe 'tarmak::vault' do
  context 'without params' do
    let(:pre_condition) {[
      """
        class{'vault_server': cloud_provider => 'aws'}
      """
    ]}

    it do
      is_expected.to compile
      is_expected.to contain_file('/etc/systemd/system/attach-ebs-volume-vault.service').with(
        'mode' => '0644',
      )
      is_expected.to contain_file('/etc/systemd/system/ensure-ebs-volume-vault-formatted.service').with(
        'mode' => '0644',
      )
    end
  end
end
