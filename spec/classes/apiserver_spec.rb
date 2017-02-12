require 'spec_helper'

describe 'kubernetes::apiserver' do
  let :service_file do
    '/etc/systemd/system/kube-apiserver.service'
  end

  context 'with default values for all parameters' do
    it { should contain_class('kubernetes::apiserver') }
    it do
      should contain_file(service_file).with_content(/After=network.target/)
      should contain_file(service_file).with_content(/User=kubernetes/)
      should contain_file(service_file).with_content(/Group=kubernetes/)
      should contain_file(service_file).with_content(/#{Regexp.escape('--etcd-servers="http://localhost:2379"')}/)
      should contain_file(service_file).with_content(%r{--service-cluster-ip-range=10\.254\.0\.0/16})
    end

  context 'with etcd override for events' do
    let(:params) { {'etcd_events_port' => 1234 } }
    it 'should have an etcd overrides line' do
      should contain_file(service_file).with_content(/#{Regexp.escape('--etcd-servers-overrides="/events#http://localhost:1234"')}/)
    end
  end
  end
end
