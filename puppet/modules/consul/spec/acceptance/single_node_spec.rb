require 'spec_helper_acceptance'
require 'rubygems/package'

def prepare_host_files(host)
  file = Tempfile.new('params_tar')

  Gem::Package::TarWriter.new(file) do |writer|
    writer.add_file("etc/facter/facts.d/consul", 0700) do |f|
      content = <<EOS
#!/bin/bash
echo CONSUL_MASTER_TOKEN=7f0c1dae-aac7-44ea-abe8-d9411c9068cb
echo CONSUL_BOOTSTRAP_EXPECT=1
echo CONSUL_ENCRYPT=GFoaCb3cOofGJn2qwqvz8A==
EOS
      f.write(content)
    end
  end

  file.close

  rsync_to(host, file.path, "/tmp/archive.tar", {})
  on host, "tar xvf /tmp/archive.tar -C /"
end

if hosts.length == 1
  describe '::consul' do
    before(:all) do
      hosts.each do |host|
        # Ensure /opt/bin is in the path
        on host, "echo -e '# ensure /opt/bin is in the path\nif ! echo $$PATH | grep -q /opt/bin ; then\n  export PATH=$PATH:/opt/bin\nfi\n' > /etc/profile.d/opt-bin.sh"
      end
    end

    hosts.each do |host|
      prepare_host_files(host)
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

      hosts_as('consul').each do |host|
        it "test consul node output on host #{host.name}" do
          nodes = host.shell("eval \"$(cat /etc/consul/master-token) /opt/bin/consul members -detailed\"").stdout.split("\n").drop(1)
          expect(nodes.length).to eq(1)
        end
      end
    end
  end
end
