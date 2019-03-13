require 'spec_helper'

describe 'aws_ebs::mount', :type => :define do
  let(:title) do
    'test1'
  end

  let(:params) do
    {
      'volume_id' => 'vol-deadbeef',
      'dest_path' => '/mnt/folder/test1',
      'device' => '/dev/xvdy',
      'is_not_attached' => true
    }
  end

  let(:attach_service_name) do
    "attach-ebs-volume-test1.service"
  end

  let(:format_service_name) do
    "ensure-ebs-volume-test1-formatted.service"
  end

  let(:mount_service_name) do
    "mnt-folder-test1.mount"
  end

  context 'standard mount' do
    it 'should compile' do
      is_expected.to compile
    end

    it 'should contain attach service' do
      should contain_service(attach_service_name)
      should contain_file("/etc/systemd/system/#{attach_service_name}").with_content(%r{/opt/bin/aws_ebs_attach_volume.sh /dev/xvdy vol-deadbeef})
    end

    it 'should contain format service' do
      should contain_service(format_service_name).that_requires("Service[#{attach_service_name}]")
      should contain_file("/etc/systemd/system/#{format_service_name}").with_content(%r{/opt/bin/aws_ebs_ensure_volume_formatted.sh /dev/xvdy})
    end

    it 'should contain mount service' do
      should contain_service(mount_service_name).that_requires("Service[#{format_service_name}]")
      should contain_file("/etc/systemd/system/#{mount_service_name}").with_content(%r{Type=auto})
      should contain_file("/etc/systemd/system/#{mount_service_name}").with_content(%r{Where=/mnt/folder/test1})
      should contain_file("/etc/systemd/system/#{mount_service_name}").with_content(%r{What=/dev/xvdy})
    end
  end

  context 'make sure two mounts don\'t conflict' do
    let(:pre_condition) {[
      """
        aws_ebs::mount{'test2':
          volume_id => 'vol-beefdead',
          dest_path => '/mnt/fodeler/test2',
          device => '/dev/xvdx',
          is_not_attached => true,
        }
      """
    ]}
    it 'should compile' do
      is_expected.to compile
    end
  end
end
