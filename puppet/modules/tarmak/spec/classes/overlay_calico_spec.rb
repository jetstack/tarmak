require 'spec_helper'

describe 'tarmak::overlay_calico' do
  let(:pre_condition) {[
    """
class{'vault_client': token => 'test-token'}
class{'tarmak': role => 'master'}
include kubernetes::master
"""
  ]}

  context 'without params' do
    it do
      should contain_class('calico')
      should contain_class('calico::config')
      should contain_class('calico::node')
      should contain_class('calico::policy_controller')
      should contain_vault_client__cert_service('etcd-overlay').with_run_exec('true')
    end
  end

  context 'with service_ensure => stopped' do
    let(:pre_condition) {[
      """
class{'vault_client': token => 'test-token'}
class{ 'tarmak':
  role           => 'master',
  service_ensure => 'stopped',
}
include kubernetes::master
"""
    ]}

    it do
      should contain_class('calico')
      should contain_class('calico::config')
      should contain_class('calico::node')
      should contain_class('calico::policy_controller')
      should contain_vault_client__cert_service('etcd-overlay').with_run_exec('false')
    end
  end
end
