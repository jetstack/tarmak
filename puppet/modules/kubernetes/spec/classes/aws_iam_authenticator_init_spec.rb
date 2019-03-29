require 'spec_helper'

describe 'kubernetes::aws_iam_authenticator_init' do
  context 'disabled' do
    let(:pre_condition) {[
      """
      class{'kubernetes::aws_iam_authenticator_init':
        auth_token_webhook_file => '',
        file_ensure             => 'absent',
      }
      """
    ]}
    it { should contain_file('/etc/systemd/system/aws-iam-authenticator-init.service').with('ensure' => 'absent')}
  end
  context 'auth token webhook file' do
    let(:pre_condition) {[
      """
      class{'kubernetes::aws_iam_authenticator_init': auth_token_webhook_file => '/foo/bar/baz'}
      """
    ]}
    it { should contain_file('/etc/systemd/system/aws-iam-authenticator-init.service').with_content(/#{Regexp.escape('/foo/bar/baz')}/)
}
  end
end
