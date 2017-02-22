require 'spec_helper'
describe 'calico::policy_controller' do
  let(:pre_condition) do
    "
      class{'calico':
       etcd_cluster => [#{etcd_cluster}],
       tls => #{tls},
      }
      class kubernetes{}
      define kubernetes::apply(
        $manifests,
      ){}
    "
  end

  let(:tls) { 'false' }

  let(:etcd_cluster) { "'etcd-1'" }

  let(:manifests) do
    catalogue.resource('Kubernetes::Apply', 'calico-policy-controller').send(:parameters)[:manifests]
  end

  context 'with defaults' do

    it do
      should contain_class('calico::policy_controller')
      expect(manifests[0]).to match(%{^[-\s].*etcd_endpoints: "http://etcd-1:2359"$})
      expect(manifests[1]).to match(%{^[-\s].*name: calico-config$})
    end
  end

  context 'with custom version, multiple etcd, tls' do
    let(:etcd_cluster) { "'etcd-1', 'etcd-2', 'etcd-3'" }
    let(:tls) { 'true' }

    let(:params) do
      {
        :etcd_cert_path => '/opt/etc/etcd/tls',
        :policy_controller_version => 'v5.6.7.8'
      }
    end

    it do
      should contain_class('calico::policy_controller')
      expect(manifests[0]).to match(%{^[-\s].*etcd_endpoints: "https://etcd-1:2359,https://etcd-2:2359,https://etcd-3:2359"$})
    end

    it do
      expect(manifests[1]).to match(/^[-\s].*name: calico-config$/)
      expect(manifests[1]).to match(/^[-\s].*image: calico\/kube-policy-controller:v5.6.7.8$/)
      expect(manifests[1]).to match(/^[-\s].*mountPath: \/opt\/etc\/etcd\/tls$/)
    end
  end
end
