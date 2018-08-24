require 'spec_helper'

describe 'tarmak::etcd' do
  let(:pre_condition) {[
    """
class{'vault_client': token => 'test-token'}
"""
  ]}
  let(:facts) {{
    :hostname => 'etcd-1'
  }}

  let :attatch_volume_service do
    '/etc/systemd/system/attach-ebs-volume-etcd-data.service'
  end

  let :volume_mount do
    '/etc/systemd/system/var-lib-etcd.mount'
  end

  let :volume_formatted do
    '/etc/systemd/system/ensure-ebs-volume-etcd-data-formatted.service'
  end

  context 'without params' do
    it do
      is_expected.to compile
    end
  end

  context 'on aws' do
    let(:facts) {{
      :hostname         => 'etcd-1',
      :tarmak_volume_id => 'vol-deadcafe',
    }}

    let(:pre_condition) {[
      """
        class{'vault_client': token => 'test-token'}
        class{'tarmak': cloud_provider => 'aws'}
      """
    ]}

    it do
      is_expected.to compile
    end
  end

  context '3 node etcd cluster with start index 0' do
    let(:pre_condition) {[
      """
        class{'vault_client': token => 'test-token'}
        class{'tarmak': etcd_start_index => 0}
      """
    ]}

    context 'on node etcd-0' do
      let(:facts) {{
        :hostname => 'etcd-0'
      }}
      it do
        is_expected.to compile
      end
    end

    context 'on node etcd-3' do
      let(:facts) {{
        :hostname => 'etcd-3'
      }}
      it do
        is_expected.to compile.and_raise_error(/is not within the etcd_cluster/)
      end
    end
  end

  context '3 node etcd cluster with start index 1' do
    let(:pre_condition) {[
      """
        class{'vault_client': token => 'test-token'}
        class{'tarmak':}
      """
    ]}

    context 'on node etcd-0' do
      let(:facts) {{
        :hostname => 'etcd-0'
      }}
      it do
        is_expected.to compile.and_raise_error(/is not within the etcd_cluster/)
      end
    end

    context 'on node etcd-1' do
      let(:facts) {{
        :hostname => 'etcd-1'
      }}
      it do
        is_expected.to compile
      end
    end

    context 'etcd mount device xvd' do
      let(:facts) {{
        :hostname => 'etcd-1',
        :tarmak_volume_id => 'vol-deadcafe',
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
          class{'vault_client': token => 'test-token'}
          class{'tarmak': cloud_provider => 'aws'}
        """
      ]}

      it do
        is_expected.to compile
        should contain_file(attatch_volume_service).with_content(
        /ExecStart=\/opt\/bin\/aws_ebs_attach_volume.sh \/dev\/xvdd vol-deadcafe\n/)
        should contain_file(volume_mount).with_content(
        /What=\/dev\/xvdd\n/)
        should contain_file(volume_formatted).with_content(
            /ExecStart=\/opt\/bin\/aws_ebs_ensure_volume_formatted.sh \/dev\/xvdd\n/)
      end
    end

    context 'etcd mount device nvme' do
      let(:facts) {{
        :hostname => 'etcd-1',
        :tarmak_volume_id => 'vol-deadcafe',
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
          class{'vault_client': token => 'test-token'}
          class{'tarmak': cloud_provider => 'aws'}
        """
      ]}

      it do
        is_expected.to compile
        should contain_file(attatch_volume_service).with_content(
        /ExecStart=\/opt\/bin\/aws_ebs_attach_volume.sh \/dev\/nvme1n1 vol-deadcafe\n/)
        should contain_file(volume_mount).with_content(
        /What=\/dev\/nvme1n1\n/)
        should contain_file(volume_formatted).with_content(
            /ExecStart=\/opt\/bin\/aws_ebs_ensure_volume_formatted.sh \/dev\/nvme1n1\n/)
      end
    end
  end
end
