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


  context 'feature gates' do
    context 'without given feature gates and not enabled pod priority' do
      let(:pre_condition) {[
        """
        class{'kubernetes': scheduler_feature_gates => []}
        """
      ]}
      it 'should have default feature gates' do
        should_not contain_file(service_file).with_content(/#{Regexp.escape('--feature-gates=')}/)
      end
    end

    context 'without given feature gates and enabled pod priority' do
      let(:pre_condition) {[
        """
        class{'kubernetes': enable_pod_priority => true}
        """
      ]}
      let(:version) { '1.6.0' }
      it 'should have default feature gates' do
        should contain_file(service_file).with_content(/#{Regexp.escape('--feature-gates=PodPriority=true')}/)
      end
    end

    context 'with given feature gates' do
      let(:pre_condition) {[
        """
        class{'kubernetes': scheduler_feature_gates => ['foo=true', 'bar=true']}
        """
      ]}
      it 'should have custom feature gates' do
        should contain_file(service_file).with_content(/#{Regexp.escape('--feature-gates=foo=true,bar=true')}/)
      end
    end
  end
end
