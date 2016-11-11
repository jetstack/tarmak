require 'beaker-rspec'

# Not needed for this example as our docker files have puppet installed already
#hosts.each do |host|
#  # Install Puppet #  install_puppet
#end

RSpec.configure do |c|
  # Project root
  proj_root = File.expand_path(File.join(File.dirname(__FILE__), '..'))

  # Readable test descriptions
  c.formatter = :documentation

  # Configure all nodes in nodeset
  c.before :suite do
    # Install module and dependencies
    puppet_module_install(:source => proj_root, :module_name => 'vault_client')
    hosts.each do |host|
      # Needed for the consul module to download the binary per the modulefile
      on host, puppet('module', 'install', 'puppetlabs-stdlib'), { :acceptable_exit_codes => [0,1] }
      on host, puppet('module', 'install', 'nanliu/staging'), { :acceptable_exit_codes => [0,1] }
    end
  end
end
