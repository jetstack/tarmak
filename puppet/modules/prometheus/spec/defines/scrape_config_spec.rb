require 'spec_helper'

describe 'prometheus::scrape_config', :type => :define do
  let(:pre_condition) {[
    'include kubernetes::apiserver',
  ]}

  context 'test scrape static_configs definition' do
    let(:title) do
      'etcd_k8s'
    end

    let(:etcd_cluster_exporters) { ['etcd-exporters.example.tarmak.local'] }
    let :params do
      {
        :config => {
          'metrics_path' => '/probe',
          'params' => { 'module' => ['k8s_proxy']},
          'dns_sd_config' => [ 'names' => etcd_cluster_exporters ],
          'relabel_configs' => [{
            'source_labels' => [],
            'regex' => '(.*)',
            'target_label' => '__param_target',
            'replacement' => 'https://127.0.0.1:1234/metrics',
           }],
        },
        :order             => 02,
      }
    end

    it do
      should contain_concat__fragment("kubectl-apply-prometheus-scrape-config-etcd_k8s")
        .with_content(/- names:/)
        .with_content(/  - etcd-exporters.example.tarmak.local/)
    end
  end

  context 'test scrape kubernetes_sd_configs definition' do
    let(:title) do
      'kubernetes-apiservers'
    end
    let :params do
      {
        :config => { 'kubernetes_sd_configs' => [ "role" => "endpoints" ]},
        :order                 => 02,
      }
    end
    it do
      should contain_concat__fragment("kubectl-apply-prometheus-scrape-config-kubernetes-apiservers")
        .with_content(/- role: endpoints/)
    end
  end
end
