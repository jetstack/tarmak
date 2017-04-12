require 'spec_helper'
describe 'calico' do
  let(:pre_condition) do
    "
      class kubernetes{}
      define kubernetes::apply(
        $manifests,
      ){}
    "
  end

  let(:config_yaml) do
    catalogue.resource('Kubernetes::Apply', 'calico-config').send(:parameters)[:manifests][0]
  end

  let(:policy_controller_yaml) do
    catalogue.resource('Kubernetes::Apply', 'calico-policy-controller').send(:parameters)[:manifests][0]
  end

  let(:node_yaml) do
    catalogue.resource('Kubernetes::Apply', 'calico-node').send(:parameters)[:manifests][0]
  end

  context 'on cloud_provider aws' do
    let(:params) {
      {
        :cloud_provider => 'aws'
      }
    }

    it do
      should contain_class('calico')
      should contain_class('calico::disable_source_destination_check')
      should contain_file('/opt/bin/disable-source-destination-check.sh').with({
        'ensure' => 'file',
        'mode'   => '0755',
      })
    end
  end

  context 'on storage backend etcd' do
    let(:params) {
      {
        :backend         => 'etcd',
        :etcd_cluster    => ['etcd-1', 'etcd-2'],
      }
    }

    it 'should contain policy_controller' do
      should contain_class('calico::policy_controller')
    end

    context 'without tls' do
      it do
        expect(config_yaml).to match(%r{"http://etcd-1:2359,http://etcd-2:2359"})
        expect(config_yaml).not_to match(%r{"etcd_ca: "})
        expect(config_yaml).not_to match(%r{"etcd_cert: "})
        expect(config_yaml).not_to match(%r{"etcd_key: "})
        expect(policy_controller_yaml).not_to match(%r{path: /my/etcd-secrets})
        expect(policy_controller_yaml).not_to match(%r{mountPath: /my/etcd-secrets})
        expect(policy_controller_yaml).not_to match(%r{ETCD_CERT_FILE})
        expect(policy_controller_yaml).not_to match(%r{ETCD_KEY_FILE})
        expect(policy_controller_yaml).not_to match(%r{ETCD_CA_CERT_FILE})
      end
    end

    context 'with tls' do
      let(:params) {
        {
          :backend => 'etcd',
          :etcd_cluster    => ['etcd-1', 'etcd-2'],
          :etcd_key_file => '/my/etcd-secrets/etcd-key.pem',
          :etcd_cert_file => '/my/etcd-secrets/etcd-cert.pem',
          :etcd_ca_file => '/my/etcd-secrets/etcd-ca.pem',
        }
      }
      it do
        expect(config_yaml).to match(%r{"https://etcd-1:2359,https://etcd-2:2359"})
        expect(config_yaml).to match(%r{etcd_ca: "/my/etcd-secrets/etcd-ca\.pem"})
        expect(config_yaml).to match(%r{etcd_key: "/my/etcd-secrets/etcd-key\.pem"})
        expect(config_yaml).to match(%r{etcd_cert: "/my/etcd-secrets/etcd-cert\.pem"})
        expect(policy_controller_yaml).to match(%r{path: /my/etcd-secrets})
        expect(policy_controller_yaml).to match(%r{mountPath: /my/etcd-secrets})
      end
    end
  end
end
