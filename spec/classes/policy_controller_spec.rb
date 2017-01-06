require 'spec_helper'
describe 'calico::policy_controller' do
  context 'with defaults' do
    let(:pre_condition) { "class calico { $etcd_cluster = ['etcd1'] }" }

    it do
      should contain_class('calico::policy_controller')
      should contain_file('/root/calico-config.yaml').with_content(/^[-\s].*etcd_endpoints: \"http:\/\/etcd1:2359"$/)
      should contain_file('/root/policy-controller-deployment.yaml').with_content(/^[-\s].*name: calico-config$/)
    end
  end

  context 'with custom version, multiple etcd, tls' do
    let(:pre_condition) {[ 
      "class calico { $etcd_cluster = ['etcd1','etcd2','etcd3'] }",
      "class calico { $tls = true }"
    ]}
    let(:params) {
      {
        :etcd_cert_path => '/opt/etc/etcd/tls',
        :policy_controller_version => 'v5.6.7.8'
      }
    }

    it do
      should contain_class('calico::policy_controller')
      should contain_file('/root/calico-config.yaml').with_content(/^[-\s].*etcd_endpoints: \"https:\/\/etcd1:2359,https:\/\/etcd2:2359,https:\/\/etcd3:2359"$/)
    end

    it do
      should contain_file('/root/policy-controller-deployment.yaml').with_content(/^[-\s].*name: calico-config$/)
      should contain_file('/root/policy-controller-deployment.yaml').with_content(/^[-\s].*image: calico\/kube-policy-controller:v5.6.7.8$/)
      should contain_file('/root/policy-controller-deployment.yaml').with_content(/^[-\s].*mountPath: \/opt\/etc\/etcd\/tls$/)
    end
  end
end
