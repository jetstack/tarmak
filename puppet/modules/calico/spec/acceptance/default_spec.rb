require 'spec_helper_acceptance'
if hosts.length == 1
  describe '::calico' do
    context 'test applying the module alone' do
      # Using puppet_apply as a helper
      it 'should work once with no errors based on the example' do
        pp = <<-EOS
class { 'calico':
  etcd_cluster   => [ 'etcd1' ],
  cloud_provider => '',
}
        EOS

        # Run it once
        apply_manifest(pp, :catch_failures => true)
        #expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero
      end
    end
  end
end
