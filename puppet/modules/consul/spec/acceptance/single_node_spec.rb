require 'spec_helper_acceptance'

if hosts.length == 1
  describe '::consul' do
    before(:all) do
      hosts.each do |host|
        # Ensure /opt/bin is in the path
        on host, "echo -e '# ensure /opt/bin is in the path\nif ! echo $$PATH | grep -q /opt/bin ; then\n  export PATH=$PATH:/opt/bin\nfi\n' > /etc/profile.d/opt-bin.sh"
      end
    end

    context 'test single node consul cluster' do
      # Using puppet_apply as a helper
      it 'should work with no errors based on the example' do
        pp = <<-EOS
class{'consul':}
        EOS

        # Run it twice and test for idempotency
        apply_manifest(pp, :catch_failures => true)
        expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero
      end
    end
  end
end
