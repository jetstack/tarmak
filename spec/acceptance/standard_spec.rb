require 'spec_helper_acceptance'

describe '::etcdt' do

  context 'etcd v3' do
    # Using puppet_apply as a helper
    it 'should work with no errors based on the example' do
      pp = <<-EOS
etcd::instance{'k8s-events':
  version => '3.0.15',
}
EOS

      # Run it twice and test for idempotency
      apply_manifest(pp, :catch_failures => true)
      expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero
    end
  end
end
