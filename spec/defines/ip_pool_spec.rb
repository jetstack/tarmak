require 'spec_helper'
describe 'calico::ip_pool' do
  context 'with ip_pool, mask and ipip_enabled' do
    let(:pre_condition) {[
      "class calico { $etcd_cluster = ['etcd1'] }"
    ]}
    let(:title) { 'test' }
    let(:params) {
      {
        :ip_pool      => '10.10.0.0',
        :ip_mask      => 24,
        :ipip_enabled => 'true'
      }
    }
    it do
      should contain_file('/etc/calico/ipPool-10.10.0.0-24.yaml').with_content(/^\s.*cidr: 10.10.0.0\/24$/)
      should contain_file('/etc/calico/ipPool-10.10.0.0-24.yaml').with_content(/^\s.*ipip:\n\s.*enabled: true$/)
    end

    it do
      should contain_exec('Configure calico ipPool for CIDR 10.10.0.0-24').with({
        'user'    => 'root',
        'command' => '/usr/local/sbin/calico_helper.sh apply /etc/calico/ipPool-10.10.0.0-24.yaml',
        'unless'  => '/usr/local/sbin/calico_helper.sh get /etc/calico/ipPool-10.10.0.0-24.yaml | /usr/bin/grep 10.10.0.0/24'
      })
    end
  end
end

