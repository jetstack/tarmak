require 'spec_helper'

describe 'kubernetes::rbac' do
  let :crb_system_node_file do
    '/etc/kubernetes/apply/puppernetes-rbac-system-node.yaml'
  end

  context 'with RBAC' do
    context 'disabled' do
      let(:pre_condition) {[
        """
        class{'kubernetes': authorization_mode => ['ABAC']}
        class{'kubernetes::master':}
        """
      ]}
      it { should_not contain_file(crb_system_node_file) }
    end

    context 'enabled' do
      let(:pre_condition) {[
        """
        class{'kubernetes': authorization_mode => ['RBAC']}
        class{'kubernetes::master':}
        """
      ]}
      it { should contain_file(crb_system_node_file).with_content(%r{system:node}) }
    end
  end
end
