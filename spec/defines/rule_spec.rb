require 'spec_helper'

describe 'prometheus::rule', :type => :define do
  let(:pre_condition) {[
    'include kubernetes::apiserver',
  ]}
  let(:title) do
    'cpu-usage'
  end

  let :params do
    {
      :expr        => '(100 - (avg by (instance) (irate(node_cpu{name="node-exporter",mode="idle"}[5m])) * 100)) > 75',
      :for         => "2m",
      :summary     => '{{$labels.instance}}: High CPU usage detected',
      :description => '{{$labels.instance}}: CPU usage is above 75% (current value is: {{ $value }})',
    }
  end

  it do
    should contain_concat__fragment("kubectl-apply-prometheus-rules-cpu-usage")
      .with_content(/cpu-usage.yaml/)
      .with_content(/alert: cpu-usage/)
      .with_content(/severity: page/)
      .with_content(/"{{\$labels.instance}}: High CPU usage detected"/)
  end

  context 'specified alert_label severity' do
    let :params do
      {
        :expr => '(100 - (avg by (instance) (irate(node_cpu{name="node-exporter",mode="idle"}[5m])) * 100)) > 75',
        :for         => "2m",
        :summary     => '{{$labels.instance}}: High CPU usage detected',
        :description => '{{$labels.instance}}: CPU usage is above 75% (current value is: {{ $value }})',
        :labels      => {
          'severity' => 'critical'
        },
      }
    end
    it do
      should contain_concat__fragment("kubectl-apply-prometheus-rules-cpu-usage")
        .with_content(/cpu-usage.yaml/)
        .with_content(/alert: cpu-usage/)
        .with_content(/severity: critical/)
        .with_content(/"{{\$labels.instance}}: High CPU usage detected"/)
    end
  end
end
