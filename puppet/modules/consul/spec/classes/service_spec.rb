require 'spec_helper'

describe 'consul::service' do
    let(:pre_condition) do
        [
            'include consul',
            'include consul::install'
        ]
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
    end

    context 'mount with xvd' do
        let(:facts) {{
            :disks => {
                'xvdd' => {
                    'size' => '10.00 GiB',
                },
                'xvda' => {
                    'size' =>'32.00 GiB',
                }
            }
        }}

        let(:pre_condition) {[
          """
            class{'consul': cloud_provider => 'aws', volume_id => 'vol-deadcafe'}
          """
        ]}

        it 'should create consul mount' do

            should contain_file(systemd_dir+'/var-lib-consul.mount').with(
                :ensure => 'file',
                :mode => '0644',
            ).with_content(/What=\/dev\/xvdd\n/)
            should contain_service('attach-ebs-volume-consul.service').with(
                :ensure => 'running',
                :enable => true,
            )

            should contain_file(systemd_dir+'/attach-ebs-volume-consul.service').with(
                :ensure => 'file',
                :mode => '0644',
            ).with_content(/ExecStart=\/usr\/local\/bin\/aws_ebs_attach_volume.sh \/dev\/xvdd vol-deadcafe\n/)
            should contain_service('attach-ebs-volume-consul.service').with(
                :ensure => 'running',
                :enable => true,
            )

            should contain_file(systemd_dir+'/ensure-ebs-volume-consul-formatted.service').with(
                :ensure => 'file',
                :mode => '0644',
            ).with_content(/ExecStart=\/usr\/local\/bin\/aws_ebs_ensure_volume_formatted.sh \/dev\/xvdd\n/)
            should contain_service('ensure-ebs-volume-consul-formatted.service').with(
                :ensure => 'running',
                :enable => true,
            )
        end
    end

    context 'mount with nvme' do
        let(:facts) {{
            :disks => {
                'nvme0n1' => {
                    'size' => '32.00 GiB',
                },
                'nvme1n1' => {
                    'size' => '50.00 GiB',
                },
                'nvme2n1' => {
                    'size' => '139.70 GiB',
                }
            }
        }}

        let(:pre_condition) {[
          """
            class{'consul': cloud_provider => 'aws', volume_id => 'vol-deadcafe'}
          """
        ]}

        it 'should create consul mount' do

            should contain_file(systemd_dir+'/var-lib-consul.mount').with(
                :ensure => 'file',
                :mode => '0644',
            ).with_content(/What=\/dev\/nvme1n1\n/)
            should contain_service('attach-ebs-volume-consul.service').with(
                :ensure => 'running',
                :enable => true,
            )

            should contain_file(systemd_dir+'/attach-ebs-volume-consul.service').with(
                :ensure => 'file',
                :mode => '0644',
            ).with_content(/ExecStart=\/usr\/local\/bin\/aws_ebs_attach_volume.sh \/dev\/nvme1n1 vol-deadcafe\n/)
            should contain_service('attach-ebs-volume-consul.service').with(
                :ensure => 'running',
                :enable => true,
            )

            should contain_file(systemd_dir+'/ensure-ebs-volume-consul-formatted.service').with(
                :ensure => 'file',
                :mode => '0644',
            ).with_content(/ExecStart=\/usr\/local\/bin\/aws_ebs_ensure_volume_formatted.sh \/dev\/nvme1n1\n/)
            should contain_service('ensure-ebs-volume-consul-formatted.service').with(
                :ensure => 'running',
                :enable => true,
            )
        end
    end
end
