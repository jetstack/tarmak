require 'spec_helper'

describe 'tarmak::worker' do
  let(:pre_condition) {[
    """
class{'vault_client': token => 'test-token'}
"""
  ]}

  context 'without params' do
    it do
      is_expected.to compile
      should contain_vault_client__cert_service('kube-proxy').with_run_exec('true')
      should contain_vault_client__cert_service('kubelet').with_run_exec('true')
      should contain_class('kubernetes::kubelet').with_service_ensure('running')
      should contain_class('kubernetes::proxy').with_service_ensure('running')
    end
  end

  context 'with service_ensure => stopped' do
    let(:pre_condition) {[
      """
  class{'vault_client': token => 'test-token'}
  class{'tarmak': service_ensure => 'stopped'}
  """
    ]}
    
    it do
      is_expected.to compile
      should contain_vault_client__cert_service('kube-proxy').with_run_exec('false')
      should contain_vault_client__cert_service('kubelet').with_run_exec('false')
      should contain_class('kubernetes::kubelet').with_service_ensure('stopped')
      should contain_class('kubernetes::proxy').with_service_ensure('stopped')
    end
  end
end
