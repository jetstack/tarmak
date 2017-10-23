require 'spec_helper_acceptance'
if hosts.length == 1
  describe '::calico' do
    context 'applying the module with a running etcd cluster' do
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
  cloud_provider    => '',
  etcd_overlay_port => 2379,
}
        EOS

        # Run it twice
        apply_manifest(pp, :catch_failures => true)
        expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero
      end
    end
  end
end
