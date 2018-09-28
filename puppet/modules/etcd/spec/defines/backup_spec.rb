require 'spec_helper'

describe 'etcd::instance', :type => :define do

  let(:script) {
    contain_file('/opt/bin/etcd-test-backup.sh')
  }

  context 'global defined backup settings' do
    let(:pre_condition) {[
      """
        class{'etcd':
          backup_enabled => true,
          backup_bucket_prefix => 'my-bucket/my-prefix',
        }
      """
    ]}

    let(:title) { 'test' }

    context 'with no override' do
      let(:params) {
        {
          :version => '1.2.3',
        }
      }

      it 'contains backup' do
        should contain_class('etcd')
        should contain_etcd__backup('test')
      end

      it 'has correct aws command in script' do
        should script.with_content(/#{Regexp.escape('aws s3 cp "')}/)
      end
    end

    context 'with disabled backup in instance' do
      let(:params) {
        {
          :version => '1.2.3',
          :backup_enabled => false,
        }
      }

      it 'does not contains backup' do
        should contain_class('etcd')
        should_not contain_etcd__backup('test')
      end

    end

    context 'with enabled SSE in instance' do
      let(:params) {
        {
          :version => '1.2.3',
          :backup_sse => 'aws:kms',
        }
      }

      it 'does not contains backup' do
        should contain_class('etcd')
        should contain_etcd__backup('test')
      end

      it 'has correct aws command in script' do
        should script.with_content(/#{Regexp.escape('aws s3 cp --sse aws:kms "')}/)
      end

    end
  end
end
