require 'spec_helper'
describe 'calico::policy_controller' do
  let(:pre_condition) do
    "
      class{'calico':
       etcd_cluster => [#{etcd_cluster}],
       #{tls}
      }
      class kubernetes{}
      define kubernetes::apply(
        $manifests,
      ){}
    "
  end

  let(:tls) { '' }

  let(:etcd_cluster) { "'etcd-1'" }

  let(:manifests) do
    catalogue.resource('Kubernetes::Apply', 'calico-policy-controller').send(:parameters)[:manifests]
  end

  context 'with defaults' do
    it do
      should contain_class('calico::policy_controller')
      expect(manifests[0]).to match(%{^[-\s].*key: etcd_endpoints$})
      expect(manifests[0]).to match(%{^[-\s].*name: calico-config$})
      expect(manifests[0]).not_to match(%{^[-\s].*key: etcd_ca$})
    end
  end

  context 'with custom version, multiple etcd, tls' do
    let(:etcd_cluster) { "'etcd-1', 'etcd-2', 'etcd-3'" }
    let(:tls) do
      "
          etcd_key_file => '/my/etcd-secrets/etcd-key.pem',
          etcd_cert_file => '/my/etcd-secrets/etcd-cert.pem',
          etcd_ca_file => '/my/etcd-secrets/etcd-ca.pem',
      "
    end

    it do
      should contain_class('calico::policy_controller')
      expect(manifests[0]).to match(%{^[-\s].*key: etcd_ca$})
    end

    it do
      expect(manifests[0]).to match(/^[-\s].*name: calico-config$/)
      expect(manifests[0]).to match(/^[-\s].*image: quay.io\/calico\/kube-policy-controller:v/)
      expect(manifests[0]).to match(/^[-\s].*mountPath: \/my\/etcd-secrets$/)
    end
  end
end
