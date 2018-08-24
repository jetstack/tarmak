require 'spec_helper'

describe 'site_module::docker_storage' do

  let :conf_file do
    '/etc/sysconfig/docker-storage-setup'
  end

  context 'with default values for all parameters' do
    it {
      should contain_class('site_module::docker_storage')
      should contain_file(conf_file).with_content(/DEVS=\n/)
    }
  end

  context 'ebs device for xvd' do
    let :facts do {
        :disks => {
          'xvdd' => {
            'size' => '10.00 GiB',
          },
          'xvda' => {
            'size' =>'32.00 GiB',
          }
        }
      }
    end

    it do
      should contain_file(conf_file).with_content(/DEVS=xvdd\n/)
    end
  end

  context 'ebs device for nvme' do
    let :facts do {
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
      }
    end

    it do
      should contain_file(conf_file).with_content(/DEVS=nvme1n1\n/)
    end
  end
end
