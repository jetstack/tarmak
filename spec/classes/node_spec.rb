require 'spec_helper'
describe 'calico::node' do
  context 'with defaults, no tls, and a single node etcd cluster' do

    let(:params) {
      {
        :etcd_endpoints  => 'http://etcd1:2359',
        :etcd_cert_file  => '/etc/etcd/ssl/etcd-overlay.pem',
        :etcd_key_file   => '/etc/etcd/ssl/etcd-overlay-key.pem',
        :etcd_ca_file    => '/etc/etcd/ssl/etcd-overlay-ca.pem',
        :aws_filter_hack => true,
        :tls             => false,
      }
    }

    it do
      should contain_class('calico::node')
      should contain_file('/opt/cni/bin/calicoctl').with(
        'mode' => '0755',
      )
      should contain_file('/etc/calico/calico.env').with_content(/^ETCD_ENDPOINTS=\"http:\/\/etcd1:2359\"$/)
      should contain_file('/etc/systemd/system/calico-node.service').with(
        'notify' => '["Exec[calico-systemctl-daemon-reload]"]'
      )
      should contain_file('/etc/systemd/system/calico-node.service').with_content(/^EnvironmentFile=\/etc\/calico\/calico.env$/)
      should contain_file('/usr/local/sbin/calico_filter_hack.sh')
    end
  end

  context 'custom version, tls on, filter hack off, cert paths' do

    let(:params) {
      {
        :node_version    => 'v2.3.4',
        :aws_filter_hack => false,
        :tls             => true,
        :etcd_endpoints  => 'https://etcd1:2345,https://etcd2:2345,https://etcd3:2345',
        :etcd_cert_path  => '/opt/etc/etcd/tls',
        :etcd_cert_file  => '/opt/etc/etcd/tls/etcd-overlay.pem',
        :etcd_key_file   => '/opt/etc/etcd/tls/etcd-overlay-key.pem',
        :etcd_ca_file    => '/opt/etc/etcd/tls/etcd-overlay-ca.pem',
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
      should contain_file('/etc/calico/calico.env').with_content(/^ETCD_ENDPOINTS=\"https:\/\/etcd1:2345,https:\/\/etcd2:2345,https:\/\/etcd3:2345"$/)
    end

    it do
      should contain_exec('calico-systemctl-daemon-reload').with({
        'command'     => '/usr/bin/systemctl daemon-reload',
        'refreshonly' => 'true',
      })
      should contain_file('/etc/systemd/system/calico-node.service').with(
        'notify' => '["Exec[calico-systemctl-daemon-reload]"]'
      )
      should contain_file('/etc/systemd/system/calico-node.service').with_content(/^EnvironmentFile=\/etc\/calico\/calico.env$/)
      should contain_file('/etc/systemd/system/calico-node.service').with_content(/^ calico\/node:v2.3.4$/)
      should contain_file('/etc/systemd/system/calico-node.service').with_content(/^ -v \/opt\/etc\/etcd\/tls:\/etc\/etcd\/ssl \\$/)
    end

    it do
      should_not contain_file('/usr/local/sbin/calico_filter_hack.sh')
    end
  end
end
