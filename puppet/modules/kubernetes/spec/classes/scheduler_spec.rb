require 'spec_helper'

describe 'kubernetes::scheduler' do
  let :service_file do
    '/etc/systemd/system/kube-scheduler.service'
  end

  context 'with default values for all parameters' do
    it { should contain_class('kubernetes::scheduler') }
    it do
      should contain_file(service_file).with_content(/User=kubernetes/)
      should contain_file(service_file).with_content(/Group=kubernetes/)
      should contain_file(service_file).with_content(/#{Regexp.escape('--kubeconfig=/etc/kubernetes/kubeconfig-scheduler')}/)
      should contain_file(service_file).with_content(%r{--leader-elect=true})
    end
  end
end
