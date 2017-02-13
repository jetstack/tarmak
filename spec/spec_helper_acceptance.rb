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
        on host, 'yum install -y rsync'
      end
      logger.notify "ensure rsync exists on #{host}"
      install_dev_puppet_module_on(host, :source => module_root, :module_name => 'puppernetes', :target_module_path => $module_path)
      rsync_to(host, "#{module_root}/spec/fixtures/modules/", $module_path, {})
      on host, "chown -R 0:0 #{$module_path}"
    end
  end
end
