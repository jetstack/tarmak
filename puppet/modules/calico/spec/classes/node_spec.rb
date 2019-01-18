require 'spec_helper'
require 'yaml'

describe 'calico::node' do
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
        $config_dir = '/etc/kubernetes'
        $group = 'kubernetes'
        $apply_dir = '/etc/kubernetes/apply'
        $_apiserver_insecure_port = 1000
      }
      define kubernetes::apply(
      $manifests,
      ){
        kubernetes::addon_manager_labels($manifests[0])
      }
      class{'calico':
        #{mtu}
        #{backend}
        #{etcd_cluster}
        #{etcd_tls}
        #{typha_enabled}
      }
    "
  end

  let(:mtu) { '' }
  let(:backend) { '' }
  let(:etcd_cluster) { '' }
  let(:etcd_tls) { '' }
  let(:typha_enabled) { '' }

  let(:calico_node) do
    catalogue.resource('Kubernetes::Apply', 'calico-node').send(:parameters)[:manifests][0].join("\n")
  end

  context 'with default parameters' do
    it 'is valid yaml' do
      YAML.load(calico_node)
    end

    it 'has default mtu' do
      expect(calico_node).to match(/name: FELIX_IPINIPMTU\s+value: "1480"/)
    end
  end

  context 'with mtu 8981' do
    let(:mtu) do
      'mtu => 8981,'
    end
    it 'is valid yaml' do
      YAML.load(calico_node)
    end

    it 'has custom mtu' do
      expect(calico_node).to match(/name: FELIX_IPINIPMTU\s+value: "8981"/)
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
          expect(calico_node).to match(%r{name: etcd-certs\s+hostPath:\s+path: /etc/etcd/ssl})
          expect(calico_node).to match(%r{mountPath: /etc/etcd/ssl\s+name: etcd-certs})
        end
      end

      context 'without TLS' do
        it 'doesn\'t set up TLS' do
          expect(calico_node).not_to match(%r{name: etcd-certs\s+hostPath:\s+path: /etc/etcd/ssl})
          expect(calico_node).not_to match(%r{mountPath: /etc/etcd/ssl\s+name: etcd-certs})
        end
      end
    end
  end

  context 'with backend kubernetes' do
    let(:backend) do
      'backend => \'kubernetes\','
    end

    context 'with typha disabled' do
      let(:typha_enabled) do
          "typha_enabled => false,"
      end

      it 'has no typha' do
        expect(calico_node).not_to match(%r{name: FELIX_TYPHAK8SSERVICENAME})
      end
    end

    context 'with typha enabled' do
      let(:typha_enabled) do
          "typha_enabled => true,"
      end

      it 'has typha' do
        expect(calico_node).to match(%r{name: FELIX_TYPHAK8SSERVICENAME})
        expect(calico_node).to match(%r{name: TYPHA_CONNECTIONREBALANCINGMODE})
        expect(calico_node).to match(%r{name: TYPHA_CONNECTIONREBALANCINGMODE})
      end
    end

  end
end
