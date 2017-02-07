require 'spec_helper'

describe 'kubernetes::scheduler' do
  context 'with default values for all parameters' do
    it { should contain_class('kubernetes::scheduler') }
    it do
      have_service_file = contain_file('/etc/systemd/system/kube-scheduler.service')
      should have_service_file.with_content(/User=kubernetes/)
      should have_service_file.with_content(/Group=kubernetes/)
      should have_service_file.with_content(/#{Regexp.escape('--kubeconfig=/etc/kubernetes/kubeconfig-scheduler')}/)
    end
  end
end
