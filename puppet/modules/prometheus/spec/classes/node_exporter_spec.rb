require 'spec_helper'

describe 'prometheus::node_exporter' do
  context 'on etcd node' do
    let(:pre_condition) {[
      'class tarmak {',
      "  $role = 'etcd'",
      '  $etcd_k8s_main_client_port = 1234',
      '  $etcd_k8s_events_client_port = 1235',
      '  $etcd_overlay_client_port = 1236',
      "  $etcd_cluster_exporters = ['etcd-exporters.example.tarmak.local']",
      '}',
      'include tarmak',
    ]}

    it { should contain_class('prometheus') }
  end

  context 'on master node' do
    let(:pre_condition) {[
      'class tarmak {',
      "  $role = 'master'",
      '  $etcd_k8s_main_client_port = 1234',
      '  $etcd_k8s_events_client_port = 1235',
      '  $etcd_overlay_client_port = 1236',
      "  $etcd_cluster_exporters = ['etcd-exporters.example.tarmak.local']",
      '}',
      'include tarmak',
      'class kubernetes::apiserver{}',
      'require kubernetes::apiserver',
    ]}

    it { should contain_class('prometheus::server') }

    let :manifests_file do
      '/etc/kubernetes/apply/node-exporter.yaml'
    end

    let(:manifests) do
      catalogue.resource('Kubernetes::Apply', 'node-exporter').send(:parameters)[:manifests]
    end
    context 'with default values for all parameters' do
      it 'is valid yaml' do
        manifests.each do |manifest|
          YAML.parse manifest
        end
      end
    end
    context 'with custom port' do
      let :params do
        { :port => 1234 }
      end

      it 'should have the port set' do
        should contain_file(manifests_file).with_content(/containerPort: 1234/)
        should contain_file(manifests_file).with_content(/hostPort: 1234/)
        should contain_file(manifests_file).with_content(%r{--web.listen-address=:1234})
      end
    end

    context 'with image and version' do
      let :params do
        {
          :version => '1.2.3',
          :image   => 'prom/node',
        }
      end

      it 'should have the port set' do
        should contain_file(manifests_file).with_content(%r{image: prom/node:v1\.2\.3})
      end
    end
  end
end
