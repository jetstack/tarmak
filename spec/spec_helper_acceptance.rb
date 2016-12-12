require 'beaker-rspec'

# Install Puppet on all hosts
install_puppet_on(hosts, options)

RSpec.configure do |c|
  module_root = File.expand_path(File.join(File.dirname(__FILE__), '..'))

  c.formatter = :documentation

  c.before :suite do
    # Install module to all hosts
    hosts.each do |host|
      install_dev_puppet_module_on(host, :source => module_root, :module_name => 'vault_client', :target_module_path => '/etc/puppetlabs/code/modules')
      # Install dependencies
      on(host, puppet('module', 'install', 'puppetlabs-stdlib', '--version', '4.2.0'))
      on(host, puppet('module', 'install', 'puppet-archive', '--version', '1.1.2'))
      on(host, puppet('module', 'install', 'camptocamp-systemd', '--version', '0.4.0'))
    end
  end
end
