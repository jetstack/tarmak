require 'spec_helper'

describe 'kubernetes::apiserver' do
  context 'with default values for all parameters' do
    it { should contain_class('kubernetes::apiserver') }
    it do
      have_service_file = contain_file('/etc/systemd/system/kube-apiserver.service')
      should have_service_file.with_content(/User=kubernetes/)
      should have_service_file.with_content(/Group=kubernetes/)
      should have_service_file.with_content(/#{Regexp.escape('--etcd-servers="http://127.0.0.1:2379"')}/)
      should have_service_file.with_content(%r{--service-cluster-ip-range=10\.254\.0\.0/16})
    end
  end
end
