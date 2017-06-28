require 'spec_helper_acceptance'

case fact('osfamily')
  when 'Windows'
    command = 'cmd.exe /c echo triggered'
  else
    command = 'echo triggered'
end

describe 'with metaparameters' do
  describe 'with subscribed resources' do
    basedir = default.tmpdir('concat')

    context 'should trigger refresh' do
      pp = <<-EOS
        concat { "foobar":
          ensure => 'present',
          path   => '#{basedir}/foobar',
        }

        concat::fragment { 'foo':
          target => 'foobar',
          content => 'foo',
        }

        exec { 'trigger':
          path        => $::path,
          command     => "#{command}",
          subscribe   => Concat['foobar'],
          refreshonly => true,
        }
      EOS

      it 'applies the manifest twice with stdout regex' do
        expect(apply_manifest(pp, :catch_failures => true).stdout).to match(/Triggered 'refresh'/)
        expect(apply_manifest(pp, :catch_changes => true).stdout).to_not match(/Triggered 'refresh'/)
      end
    end
  end

  describe 'with resources to notify' do
    basedir = default.tmpdir('concat')
    context 'should notify' do
      pp = <<-EOS
        exec { 'trigger':
          path        => $::path,
          command     => "#{command}",
          refreshonly => true,
        }

        concat { "foobar":
          ensure => 'present',
          path   => '#{basedir}/foobar',
          notify => Exec['trigger'],
        }

        concat::fragment { 'foo':
          target => 'foobar',
          content => 'foo',
        }
      EOS

      it 'applies the manifest twice with stdout regex' do
        expect(apply_manifest(pp, :catch_failures => true).stdout).to match(/Triggered 'refresh'/)
        expect(apply_manifest(pp, :catch_changes => true).stdout).to_not match(/Triggered 'refresh'/)
      end
    end
  end
end
