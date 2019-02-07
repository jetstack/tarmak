require 'spec_helper'

require 'json'
require 'yaml'

describe 'calico::config' do
  let(:pre_condition) do
    "
      class kubernetes{}
      define kubernetes::apply(
      $manifests,
      ){
        kubernetes::addon_manager_labels($manifests[0])
      }
      class{'calico':
        #{cloud_provider}
        #{mtu}
        #{backend}
        #{etcd_cluster}
        #{etcd_tls}
      }
    "
  end

  let(:cloud_provider) { '' }
  let(:mtu) { '' }
  let(:backend) { '' }
  let(:etcd_cluster) { '' }
  let(:etcd_tls) { '' }

  let(:calico_config) do
    catalogue.resource('Kubernetes::Apply', 'calico-config').send(:parameters)[:manifests][0]
  end

  let (:cni_network_config) do
    y = YAML.load(calico_config)
    y['data']['cni_network_config']
  end


  context 'with default parameters' do
    it 'is valid yaml' do
      YAML.load(calico_config)
    end

    it 'has valid cni_network_config json' do
      cni = JSON.load(cni_network_config)
      expect(cni['plugins'].first['mtu']).to eq(1480)
    end
  end

  context 'with mtu 8981' do
    let(:mtu) do
      'mtu => 8981,'
    end
    it 'is valid yaml' do
      YAML.load(calico_config)
    end

    it 'has valid cni_network_config json' do
      cni = JSON.load(cni_network_config)
      expect(cni['plugins'].first['mtu']).to eq(8981)
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
          expect(calico_config).to match(%r{"https://etcd-1:2359,https://etcd-2:2359"})
          expect(calico_config).to match(%r{etcd_ca: "/my/etcd-secrets/etcd-ca\.pem"})
          expect(calico_config).to match(%r{etcd_key: "/my/etcd-secrets/etcd-key\.pem"})
          expect(calico_config).to match(%r{etcd_cert: "/my/etcd-secrets/etcd-cert\.pem"})
        end
      end

      context 'without TLS' do
        it 'doesn\'t set up TLS' do
          expect(calico_config).to match(%r{"http://etcd-1:2359,http://etcd-2:2359"})
          expect(calico_config).not_to match(%r{"etcd_ca: "})
          expect(calico_config).not_to match(%r{"etcd_cert: "})
          expect(calico_config).not_to match(%r{"etcd_key: "})
        end
      end
    end
  end

  context 'with backend kubernetes' do
    let(:backend) do
      'backend => \'kubernetes\','
    end

    it 'sets up ipam' do
      expect(calico_config).to match(%r{"type": "host-local",})
      expect(calico_config).to match(%r{"subnet": "usePodCidr"})
    end

    it do
      is_expected.to compile
    end
  end
end
