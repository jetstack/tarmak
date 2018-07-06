require 'spec_helper'

describe 'prometheus::blackbox_exporter_etcd' do
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
  end
end
