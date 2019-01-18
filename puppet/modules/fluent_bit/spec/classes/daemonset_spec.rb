require 'spec_helper'
describe 'fluent_bit::daemonset' do
  let(:pre_condition) do
    [
      'include kubernetes::apiserver'
    ]
  end

  let :manifests_file do
    '/etc/kubernetes/apply/fluent-bit.yaml'
  end

  context 'with default values for all parameters' do
    it 'should write manifests' do
      should contain_file(manifests_file).with_content(/#{Regexp.escape('*_kube-system_*,*_service-broker_*,*_monitoring_*')}/)
      should contain_file(manifests_file).with_content(/#{Regexp.escape('10.254.0.1 kubernetes.default.svc.cluster.local')}/)
    end
  end
end
