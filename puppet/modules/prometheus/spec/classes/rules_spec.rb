require 'spec_helper'

describe 'prometheus' do
  let(:pre_condition) {[
    'class tarmak {',
    "  $role = 'etcd'",
    '  $etcd_k8s_main_client_port = 1234',
    '  $etcd_k8s_events_client_port = 1235',
    '  $etcd_overlay_client_port = 1236',
    "  $etcd_cluster_exporters = ['etcd-exporters.example.tarmak.local']",
    '}',
    'include tarmak',
    'class kubernetes::apiserver{}',
    'require kubernetes::apiserver',
    "class{'prometheus': role => 'master'}",
    'include prometheus::server',
    'include prometheus::node_exporter',
    'include prometheus::blackbox_exporter_etcd',
    'include prometheus::kube_state_metrics',
  ]}

  let :rules_file do
    '/etc/kubernetes/apply/prometheus-rules.yaml'
  end

  let(:rules_manifest) do
    concat = catalogue.resource('concat_file', rules_file)
    concat.to_ral.should_content
  end

  context 'rules with default values for all classes' do
    it 'are valid yaml' do
      YAML.parse rules_manifest
    end

    it 'rules from all classes are added' do
      expect(rules_manifest).to match(/ScrapeEndpointDown/)
      expect(rules_manifest).to match(/NodeHighCPUUsage/)
      expect(rules_manifest).to match(/EtcdNoLeader/)
      expect(rules_manifest).to match(/KubernetesPodUnready/)
    end

    it 'are valid prometheus rules' do
      skip 'no promtool found in PATH' unless promtool_available?
      config = YAML.load rules_manifest
      file = Tempfile.new('prometheus-rules')
      config['data'].each do |key, value|
        file.write("---\n")
        file.write(value)
      end
      file.close
      Open3.popen3('promtool', 'check', 'rules', file.path) do |i,o,e,t|
        i.close
        expect(t.value.exitstatus).to eq(0), "validation of prometheus rules had errors: #{o.read} #{e.read}"
      end
    end
  end
end
