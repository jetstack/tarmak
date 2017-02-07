require 'spec_helper'

describe 'kubernetes::controller_manager' do
  context 'with default values for all parameters' do
    it { should contain_class('kubernetes::controller_manager') }
    it do
      have_service_file = contain_file('/etc/systemd/system/kube-controller-manager.service')
      should have_service_file.with_content(/User=kubernetes/)
      should have_service_file.with_content(/Group=kubernetes/)
      should have_service_file.with_content(%r{--kubeconfig=/etc/kubernetes/kubeconfig-controller-manager})
    end
    it do
      have_kubeconfig_file = contain_file('/etc/kubernetes/kubeconfig-controller-manager')
      should have_kubeconfig_file.with_content(%r{server: http://127\.0\.0\.1:8080})
    end
  end
end
