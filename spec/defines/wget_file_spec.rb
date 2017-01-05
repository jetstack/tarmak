require 'spec_helper'
describe 'calico::wget_file' do
  context 'with supplied download_dir and url' do
    let(:title) { 'test' }
    let(:params) {
      {
        :url             => 'https://example.com.ac.io.uk/somepath/somethingelse/file.pp',
        :destination_dir => '/opt/bin/rubbish/bin'
      }
    }
    it do
      should contain_exec('download-file.pp').with({
        'command' => '/usr/bin/wget -O file.pp https://example.com.ac.io.uk/somepath/somethingelse/file.pp',
        'cwd'     => '/opt/bin/rubbish/bin',
        'creates' => '/opt/bin/rubbish/bin/file.pp'
      }) 
    end
  end

  context 'with different filename and trailing slash' do
    let(:title) { 'test' }
    let(:params) {
      {
        :url              => 'https://example.com.ac.io.uk/somepath/somethingelse/file.pp',
        :destination_dir  => '/opt/bin/skip/',
        :destination_file => 'nail.qq'
      }
    }
    it do
      should contain_exec('download-nail.qq').with({
        'command' => '/usr/bin/wget -O nail.qq https://example.com.ac.io.uk/somepath/somethingelse/file.pp',
        'cwd'     => '/opt/bin/skip/',
        'creates' => '/opt/bin/skip//nail.qq'
      })
    end
  end
end
