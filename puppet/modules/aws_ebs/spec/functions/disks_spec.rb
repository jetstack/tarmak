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

      $exp_result = ['abc', 'cba', 'zzz']
      it { is_expected.to run.with_params().and_return($exp_result) }
    end
end
