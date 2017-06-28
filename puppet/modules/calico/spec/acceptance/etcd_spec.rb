require 'spec_helper_acceptance'
if hosts.length == 1
  describe '::calico' do
    context 'applying the module with a running etcd cluster, tls / aws turned off, doing the filter hack and creating an IP pool' do
      before(:all) do
        # use the repo version of etcd with default settings
        shell 'yum -y install etcd'
        shell 'systemctl start etcd.service'
      end
      # Using puppet_apply as a helper
      it 'should work with no errors' do
        pp = <<-EOS
class { 'calico':
  etcd_cluster      => [ 'localhost' ],
  aws               => false,
  aws_filter_hack   => true,
  tls               => false,
  etcd_overlay_port => 2379,
}
calico::ip_pool { '10.234.235.0/24':
  ip_pool      => '10.234.235.0',
  ip_mask      => 24,
  ipip_enabled => 'false',
}
        EOS

        # Run it twice
        apply_manifest(pp, :catch_failures => true)
        expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero
      end
    end
  end
end
