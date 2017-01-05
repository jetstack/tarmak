require 'spec_helper'
describe 'calico::node' do
  context 'with defaults and a single node etcd cluster' do

    let(:pre_condition) {[ 
      "class calico { $etcd_cluster = ['etcd1'] }",
      "exec {'calico-systemctl-daemon-reload':}"
    ]}

    it do
      should contain_class('calico::node')
      should contain_file('/opt/cni/bin/calicoctl').with(
        'mode' => '0755',
      )
      should contain_file('/etc/calico/calico.env').with_content(/^ETCD_ENDPOINTS=\"http:\/\/etcd1:2359\"$/)
      should contain_file('/usr/lib/systemd/system/calico-node.service').with(
        'notify' => '["Exec[calico-systemctl-daemon-reload]"]'
      )
      should contain_file('/usr/lib/systemd/system/calico-node.service').with_content(/^EnvironmentFile=\/etc\/calico\/calico.env$/)
      should contain_file('/usr/local/sbin/calico_filter_hack.sh')
    end
  end

  context 'custom version, tls on, filter hack off, custom port, cert path, multi-node etcd cluster' do

    let(:pre_condition) {[
      "class calico { $etcd_cluster = ['etcd1','etcd2','etcd3','etcd4','etcd5'] }",
      "class calico { $etcd_overlay_port = 2345 }",
      "class calico { $tls = true }",
      "exec {'calico-systemctl-daemon-reload':}"
    ]}

    let(:params) {
      {
        :node_version    => 'v2.3.4',
        :aws_filter_hack => false,
        :etcd_cert_path  => '/opt/etc/etcd/tls',
      }
    }

    it do
      should contain_class('calico::node')
      should contain_file('/opt/cni/bin/calicoctl').with(
        'mode' => '0755'
      )
      should contain_calico__wget_file('calicoctl').with(
        'url'  => 'https://github.com/projectcalico/calico-containers/releases/download/v2.3.4/calicoctl'
      )
      should contain_file('/etc/calico/calico.env').with_content(/^ETCD_ENDPOINTS=\"https:\/\/etcd1:2345,https:\/\/etcd2:2345,https:\/\/etcd3:2345,https:\/\/etcd4:2345,https:\/\/etcd5:2345\"$/)
    end

    it do
      should contain_file('/usr/lib/systemd/system/calico-node.service').with(
        'notify' => '["Exec[calico-systemctl-daemon-reload]"]'
      )
      should contain_file('/usr/lib/systemd/system/calico-node.service').with_content(/^EnvironmentFile=\/etc\/calico\/calico.env$/)
      should contain_file('/usr/lib/systemd/system/calico-node.service').with_content(/^ calico\/node:v2.3.4$/)
      should contain_file('/usr/lib/systemd/system/calico-node.service').with_content(/^ -v \/opt\/etc\/etcd\/tls:\/etc\/etcd\/ssl \\$/)
    end

    it do
      should_not contain_file('/usr/local/sbin/calico_filter_hack.sh')
    end
  end
end
