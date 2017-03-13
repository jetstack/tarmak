require 'spec_helper'

describe 'kubernetes::controller_manager' do
  let :service_file do
    '/etc/systemd/system/kube-controller-manager.service'
  end

  context 'with default values for all parameters' do
    it { should contain_class('kubernetes::controller_manager') }
    it do
      should contain_file(service_file).with_content(/User=kubernetes/)
      should contain_file(service_file).with_content(/Group=kubernetes/)
      should contain_file(service_file).with_content(%r{--kubeconfig=/etc/kubernetes/kubeconfig-controller-manager})
      should contain_file(service_file).with_content(%r{--leader-elect=true})
    end
    it do
      have_kubeconfig_file = contain_file('/etc/kubernetes/kubeconfig-controller-manager')
      should have_kubeconfig_file.with_content(%r{server: http://127\.0\.0\.1:8080})
    end
  end
end
