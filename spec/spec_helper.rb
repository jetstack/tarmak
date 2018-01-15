require 'puppetlabs_spec_helper/module_spec_helper'

module Helpers
  def which(cmd)
    exts = ENV['PATHEXT'] ? ENV['PATHEXT'].split(';') : ['']
    ENV['PATH'].split(File::PATH_SEPARATOR).each do |path|
      exts.each { |ext|
        exe = File.join(path, "#{cmd}#{ext}")
        return exe if File.executable?(exe) && !File.directory?(exe)
      }
    end
    return nil
  end

  def promtool_available?
    not which('promtool').nil?
  end
end

RSpec.configure do |config|
  config.default_facts = {
    :path => '/bin:/sbin:/usr/bin:/usr/sbin:/opt/bin',
    :osfamily => 'RedHat',
  }
  config.include Helpers
end
