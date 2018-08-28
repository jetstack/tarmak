require 'spec_helper'

describe 'consul::service' do
    let(:pre_condition) do
      """
        class{'consul': cloud_provider => 'aws' }
      """
    end

    let :systemd_dir do
        '/etc/systemd/system'
    end

    context 'with default values for all parameters' do
        it { should contain_class('consul::service') }

        it 'should create consul exporter unit' do
            should contain_file(systemd_dir+'/consul-exporter.service').with(
                :ensure => 'file',
            )
            should contain_service('consul-exporter.service').with(
                :ensure => 'running',
                :enable => true,
            )
        end

        it 'should create consul unit' do
            should contain_file(systemd_dir+'/consul.service').with(
                :ensure => 'file',
            )
            should contain_service('consul.service').with(
                :ensure => 'running',
                :enable => true,
            )
        end

        it 'should create consul mount' do
            should contain_file(systemd_dir+'/var-lib-consul.mount')
        end

        it 'should create consul backup unit' do
            should contain_file(systemd_dir+'/consul-backup.service').with(
                :ensure => 'file',
                :mode => '0644',
            )
        end

        it 'should create consul backup timer' do
            should contain_file(systemd_dir+'/consul-backup.timer').with(
                :ensure => 'file',
            )
            should contain_service('consul-backup.timer').with(
                :ensure => 'running',
                :enable => true,
            )
        end

        it 'should create consul attach ebs unit' do
            should contain_file(systemd_dir+'/attach-ebs-volume.service').with(
                :ensure => 'file',
                :mode => '0644',
            )
            should contain_service('attach-ebs-volume.service').with(
                :ensure => 'running',
                :enable => true,
            )
        end

        it 'should create consul ensure ebs volume formatted unit' do
            should contain_file(systemd_dir+'/ensure-ebs-volume-formatted.service').with(
                :ensure => 'file',
                :mode => '0644',
            )
            should contain_service('ensure-ebs-volume-formatted.service').with(
                :ensure => 'running',
                :enable => true,
            )
        end
    end
end
