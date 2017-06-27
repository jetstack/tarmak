require 'spec_helper'

describe 'kubernetes::rbac' do
  let :crb_system_node_file do
    '/etc/kubernetes/apply/puppernetes-rbac.yaml'
  end

  context 'with RBAC' do
    context 'disabled in 1.6' do
      let(:pre_condition) {[
        """
        class{'kubernetes': authorization_mode => ['ABAC'], version => '1.6.4' }
        class{'kubernetes::master':}
        """
      ]}
      it { should_not contain_file(crb_system_node_file) }
    end

    context 'enabled in 1.6' do
      let(:pre_condition) {[
        """
        class{'kubernetes': authorization_mode => ['ABAC'], version => '1.6.4' }
        class{'kubernetes::master':}
        """
      ]}
      it { should_not contain_file(crb_system_node_file) }
    end

    context 'enabled in 1.5' do
      let(:pre_condition) {[
        """
        class{'kubernetes': authorization_mode => ['RBAC'], version => '1.5.7' }
        class{'kubernetes::master':}
        """
      ]}
      it { should contain_file(crb_system_node_file).with_content(%r{cluster-admin}) }
      it { should contain_file(crb_system_node_file).with_content(%r{system:node}) }
    end
  end
end
