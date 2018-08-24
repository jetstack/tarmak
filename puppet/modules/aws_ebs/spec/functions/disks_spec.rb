require 'spec_helper'

describe 'aws_ebs::disks' do
    context 'default values with no disks found in facts' do
      it { is_expected.to run.with_params().and_return([]) }
    end

    context 'default values with disks found in facts' do

      let :facts do {
          :disks => {
            'cba' => {
              'size' => '50.00 GiB',
            },
            'zzz' => {
              'size' => '50.00 GiB',
            },
            'abc' => {
              'size' => '50.00 GiB',
            },
          }
        }
      end

      it { is_expected.to run.with_params().and_return(['abc', 'cba', 'zzz']) }
    end

    context 'sort disks by correct logical order' do
      let :facts do {
          :disks => {
            'nvme3n5' => {
              'size' => '50.00 GiB',
            },
            'nvme0n1' => {
              'size' => '50.00 GiB',
            },
            'nvme10n1' => {
              'size' => '50.00 GiB',
            },
            'nvme2n2' => {
              'size' => '50.00 GiB',
            },
            'nvme11n2' => {
              'size' => '50.00 GiB',
            },
            'nvme2n1' => {
              'size' => '50.00 GiB',
            },
            'nvme11n1' => {
              'size' => '50.00 GiB',
            },
          }
        }
      end

      $exp_result = ['nvme0n1', 'nvme2n1', 'nvme2n2', 'nvme3n5', 'nvme10n1', 'nvme11n1', 'nvme11n2']
      it { is_expected.to run.with_params().and_return($exp_result) }

    end
end
