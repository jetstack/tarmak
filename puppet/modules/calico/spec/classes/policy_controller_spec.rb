require 'spec_helper'

describe 'calico::policy_controller' do
  let(:kubernetes_version) do
    '1.6.1'
  end

  let(:authorization_mode) do
    '[\'RBAC\']'
  end

  let(:pre_condition) do
    "
      class kubernetes{
        $_authorization_mode = #{authorization_mode}
        $version = '#{kubernetes_version}'
      }
      define kubernetes::apply(
        $manifests,
      ){}
      class{'calico':
        #{mtu}
        #{backend}
        #{etcd_cluster}
        #{etcd_tls}
      }
    "
  end

  let(:mtu) { '' }
  let(:backend) { '' }
  let(:etcd_cluster) { '' }
  let(:etcd_tls) { '' }

  let(:calico_policy_controller) do
    catalogue.resource('Kubernetes::Apply', 'calico-policy-controller').send(:parameters)[:manifests][0]
  end

  context 'with default parameters' do
    it 'is valid yaml' do
      YAML.load(calico_policy_controller)
    end
  end

  context 'with backend etcd' do
    let(:backend) do
      'backend => \'etcd\','
    end

    context 'with two node etcd cluster' do
      let(:etcd_cluster) do
        "etcd_cluster => ['etcd-1', 'etcd-2'],"
      end

      context 'with TLS' do
        let(:etcd_tls) do
          "etcd_key_file => '/my/etcd-secrets/etcd-key.pem',
          etcd_cert_file => '/my/etcd-secrets/etcd-cert.pem',
          etcd_ca_file => '/my/etcd-secrets/etcd-ca.pem',
          "
        end
        it 'sets up TLS' do
          expect(calico_policy_controller).to match(/^[-\s].*name: calico-config$/)
          expect(calico_policy_controller).to match(/^[-\s].*image: quay.io\/calico\/kube-policy-controller:v/)
          expect(calico_policy_controller).to match(/^[-\s].*mountPath: \/my\/etcd-secrets$/)
        end
      end

      context 'without TLS' do
        it 'doesn\'t set up TLS' do
          expect(calico_policy_controller).to match(%{^[-\s].*key: etcd_endpoints$})
          expect(calico_policy_controller).to match(%{^[-\s].*name: calico-config$})
          expect(calico_policy_controller).not_to match(%{^[-\s].*key: etcd_ca$})
        end
      end
    end
  end
end

