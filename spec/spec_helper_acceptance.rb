require 'beaker-rspec'

# Install Puppet on all hosts
install_puppet_on(hosts, options)

RSpec.configure do |c|
  module_root = File.expand_path(File.join(File.dirname(__FILE__), '..'))

  c.formatter = :documentation

  c.before :suite do
    # Install module to all hosts
    hosts.each do |host|
      install_dev_puppet_module_on(host, :source => module_root, :module_name => 'etcd', :target_module_path => '/etc/puppetlabs/code/modules')
      # Install dependencies
      on(host, puppet('module', 'install', 'puppetlabs-stdlib', '--version', '4.2.0'))
      on(host, puppet('module', 'install', 'camptocamp-systemd', '--version', '0.4.0'))
      #host.shell('mkdir -p /etc/puppetlabs/code/modules/vault_client')
      #host.shell('curl -sL https://github.com/puppernetes/module-vault_client/archive/master.tar.gz | tar xvzf - -C /etc/puppetlabs/code/modules/vault_client --strip-components=1')
    end
  end
end
