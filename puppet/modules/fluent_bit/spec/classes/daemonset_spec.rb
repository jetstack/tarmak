require 'spec_helper'
describe 'fluent_bit::daemonset' do
  let(:pre_condition) do
    [
      'include kubernetes::apiserver'
    ]
  end

  let :service_file do
    '/etc/systemd/system/kubectl-apply-fluent-bit.service'
  end

  let :manifests_file do
    '/etc/kubernetes/apply/fluent-bit.yaml'
  end

  context 'with default values for all parameters' do
    it 'should write systemd unit for applying' do
      should contain_file(service_file).with_content(/User=kubernetes/)
    end

    it 'should write manifests' do
      should contain_file(manifests_file).with_content(/#{Regexp.escape('*_kube-system_*,*_service-broker_*,*_monitoring_*')}/)
    end
  end
end
