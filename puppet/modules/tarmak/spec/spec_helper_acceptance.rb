require 'beaker-rspec'
require 'rspec/retry'

$module_path = '/etc/puppetlabs/code/modules/'

# Install Puppet on all hosts
install_puppet_agent_on(hosts, {
  :puppet_agent_version          => '1.9.1',
})

RSpec.configure do |c|
  module_root = File.expand_path(File.join(File.dirname(__FILE__), '..'))

  c.formatter = :documentation

  c.before :suite do
    # Sync modules to all hosts
    hosts.each do |host|
      if fact('osfamily') == 'RedHat'
        logger.notify "ensure rsync exists on #{host}"
        on host, 'yum install -y rsync'
      end

      rsync_source_path = "#{module_root}/spec/fixtures/modules"
      if File.basename(File.dirname(module_root)) == 'modules'
        rsync_source_path = File.expand_path(File.join(module_root, ".."))
        logger.notify "override rsync source to #{rsync_source_path}"
      end
      rsync_to(host, rsync_source_path, $module_path, {})
      install_dev_puppet_module_on(host, :source => module_root, :module_name => 'tarmak', :target_module_path => $module_path)
    end
  end
end
