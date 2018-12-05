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
    end
  end

  context 'cloud provider' do
    context 'default' do
      it { should_not contain_file(service_file).with_content(%r{--cloud-provider}) }
    end

    context 'aws' do
      let(:pre_condition) {[
        """
        class{'kubernetes': cloud_provider => 'aws'}
        """
      ]}
      it { should contain_file(service_file).with_content(%r{--cloud-provider=aws}) }
    end
  end

  context 'with RBAC' do
    context 'disabled' do
      let(:pre_condition) {[
        """
        class{'kubernetes': authorization_mode => ['ABAC']}
        """
      ]}
      it { should_not contain_file(service_file).with_content(%r{--use-service-account-credentials}) }
    end

    context 'enabled with kubernetes 1.6+' do
      let(:pre_condition) {[
        """
        class{'kubernetes': authorization_mode => ['RBAC'], version => '1.6.4'}
        """
      ]}
      it { should contain_file(service_file).with_content(%r{--use-service-account-credentials}) }
    end

    context 'enabled before kubernetes 1.6' do
      let(:pre_condition) {[
        """
        class{'kubernetes': authorization_mode => ['RBAC'], version => '1.5.7'}
        """
      ]}
      it { should_not contain_file(service_file).with_content(%r{--use-service-account-credentials}) }
    end
  end

  context 'feature gates' do
    context 'without given feature gates and not enabled pod priority' do
      let(:params) { {'feature_gates' => {}}}
      it 'should have default feature gates' do
        should_not contain_file(service_file).with_content(/#{Regexp.escape('--feature-gates=')}/)
      end
    end

    context 'with given feature gates' do
      let(:params) { {'feature_gates' => {'foo' => true, 'bar' => true}}}
      it 'should have custom feature gates' do
        should contain_file(service_file).with_content(/#{Regexp.escape('--feature-gates=foo=true,bar=true')}/)
      end
    end
  end
end
