require 'spec_helper'

describe 'puppernetes::etcd' do
  let(:pre_condition) {[
    """
class{'vault_client': token => 'test-token'}

class systemd{
    exec {'systemctl-daemon-reload':}
}
    """
  ]}

  context 'without params' do
    it do
      is_expected.to compile
    end
  end
end
