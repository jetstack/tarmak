require 'spec_helper'

describe 'etcd::instance', :type => :define do

  let(:pre_condition) {[
    'class systemd{}'
  ]}

  let(:config) {
    contain_file('/etc/systemd/system/etcd-test.service')
  }

  context 'single node cluster' do
    let(:title) { 'test' }
    let(:params) {
      {
        :version => '1.2.3',
        :client_port => 1234,
        :peer_port => 4321,
      }
    }

    it do
      should contain_class('etcd')
    end

    it 'should configure etcd right' do
      should config.with_content(/#{Regexp.escape('ETCD_LISTEN_PEER_URLS=http://127.0.0.1:4321')}/)
      should_not config.with_content(/^Environment=ETCD_INITIAL_/)
    end
  end

  context 'three node cluster' do
    let(:facts) {
      {
        :fqdn => 'etcd1',
      }
    }
    let(:title) { 'test' }
    let(:params) {
      {
        :version => '1.2.3',
        :members => 3,
        :client_port => 1234,
        :peer_port => 4321,
        :initial_cluster => ['etcd1','etcd2','etcd3'],
      }
    }

    it do
      should contain_class('etcd')
    end

    it 'should configure etcd right' do
      should config.with_content(/#{Regexp.escape('ETCD_LISTEN_PEER_URLS=http://0.0.0.0:4321')}/)
      should config.with_content(/#{Regexp.escape('ETCD_INITIAL_CLUSTER_TOKEN=etcd-test-7a303106eca78c4723ee70e5b0fcb891')}$/)
      should config.with_content(/#{Regexp.escape('ETCD_INITIAL_CLUSTER=etcd1=http://etcd1:4321,etcd2=http://etcd2:4321,etcd3=http://etcd3:4321')}/)
      should config.with_content(/#{Regexp.escape('ETCD_INITIAL_ADVERTISE_PEER_URLS=http://etcd1:4321')}/)
    end
  end
end
