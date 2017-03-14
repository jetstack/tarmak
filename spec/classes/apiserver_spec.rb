require 'spec_helper'

describe 'kubernetes::apiserver' do
  let :service_file do
    '/etc/systemd/system/kube-apiserver.service'
  end

  context 'with default values for all parameters' do
    it { should contain_class('kubernetes::apiserver') }
    it do
      should contain_file(service_file).with_content(/After=network.target/)
      should contain_file(service_file).with_content(/User=kubernetes/)
      should contain_file(service_file).with_content(/Group=kubernetes/)
      should contain_file(service_file).with_content(/#{Regexp.escape('"--etcd-servers=http://localhost:2379"')}/)
      should contain_file(service_file).with_content(%r{--service-cluster-ip-range=10\.254\.0\.0/16})
    end
  end

  context 'with etcd override for events' do
    let(:params) { {'etcd_events_port' => 1234 } }
    it 'should have an etcd overrides line' do
      should contain_file(service_file).with_content(/#{Regexp.escape('"--etcd-servers-overrides=/events#http://localhost:1234"')}/)
    end
  end

  context 'insecure bind address' do
    context 'is specified' do
      let(:params) { {'insecure_bind_address' => '127.0.0.1' } }
      it { should contain_file(service_file).with_content(/#{Regexp.escape('--insecure-bind-address=127.0.0.1')}/)}
    end
    context 'not specified' do
      it { should_not contain_file(service_file).with_content(/#{Regexp.escape('--insecure-bind-address=')}/)}
    end
  end

  context 'cloud provider' do
    context 'default' do
      it { should_not contain_file(service_file).with_content(%r{--cloud-provider}) }
    end

    context 'aws' do
      let(:pre_condition) {[
        """
        class{'kubernetes': cloud_provider => 'aws'}
        """
      ]}
      it { should contain_file(service_file).with_content(%r{--cloud-provider=aws}) }
    end
  end

  context 'admission controllers' do
    context 'customized' do
      let(:params) { {'admission_control' => ['Test1'] } }
      it { should contain_file(service_file).with_content(/#{Regexp.escape('--admission-control=Test1')}/)}
    end

    context 'default pre 1.4' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.3.5'}
        """
      ]}
      it { should contain_file(service_file).with_content(/#{Regexp.escape('--admission-control=NamespaceLifecycle,LimitRanger,ServiceAccount,ResourceQuota')}/)}
    end

    context 'default 1.4+' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.4.0'}
        """
      ]}
      it { should contain_file(service_file).with_content(/#{Regexp.escape('--admission-control=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,ResourceQuota')}/)}
    end
  end

  context 'storage backend' do
    context 'customized' do
      let(:params) { {'storage_backend' => 'consulator' } }
      it { should contain_file(service_file).with_content(/#{Regexp.escape('--storage-backend=consulator')}/)}
    end

    context 'default pre 1.5' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.4.8'}
        """
      ]}
      it { should_not contain_file(service_file).with_content(/#{Regexp.escape('--storage-backend=etcd2')}/)}
    end

    context 'default 1.5+' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.5.0'}
        """
      ]}
      it { should contain_file(service_file).with_content(/#{Regexp.escape('--storage-backend=etcd2')}/)}
    end
  end
end
