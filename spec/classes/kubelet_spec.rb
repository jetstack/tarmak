require 'spec_helper'

describe 'kubernetes::kubelet' do

  let :service_file do
      '/etc/systemd/system/kubelet.service'
  end

  context 'defaults' do
    it do
      should contain_file(service_file).with_content(/--register-node=true/)
      should contain_file(service_file).with_content(/--register-schedulable=true/)
      should contain_file(service_file).with_content(/--node-labels=role=worker/)
      should contain_file(service_file).with_content(/--cluster-dns=10.254.0.10/)
      should contain_file(service_file).with_content(/--cluster-domain=cluster.local/)
      should_not contain_file(service_file).with_content(/--network-plugin/)
      should contain_file(service_file).with_content(/--container-runtime=docker/)
      should contain_file(service_file).with_content(%r{--kubeconfig=/etc/kubernetes/kubeconfig-kubelet})
      should contain_file(service_file).with_content(%r{--require-kubeconfig})
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

  context 'network_plugin enabled' do
    let(:params) { {'network_plugin' => 'kubenet' } }
      it do
        should contain_file(service_file).with_content(/--network-plugin=kubenet/)
        should contain_file(service_file).with_content(/--network-plugin-mtu=1460/)
      end
  end

  context 'with role master' do
    let(:params) { {'role' => 'master' } }

    it do
      have_service_file = contain_file('/etc/systemd/system/kubelet.service')
      should have_service_file.with_content(/--register-schedulable=false/)
      should have_service_file.with_content(/--node-labels=role=master/)
    end
  end

  context 'with role worker' do
    let(:params) { {'role' => 'worker' } }

    it do
      have_service_file = contain_file('/etc/systemd/system/kubelet.service')
      should have_service_file.with_content(/--register-schedulable=true/)
      should have_service_file.with_content(/--node-labels=role=worker/)
    end
  end
  
  context 'apiservers config --api-servers vs --kubeconfig' do
    let(:params) { {'ca_file' => '/tmp/ca.pem' } }
    context 'versions before 1.4' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.3.8'}
        """
      ]}
      it do
        should_not contain_file(service_file).with_content(%r{--require-kubeconfig})
        should contain_file(service_file).with_content(%r{--api-servers=http://127\.0\.0\.1:8080})
      end
    end

    context 'versions 1.4+' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.4.0'}
        """
      ]}
      it do
        should contain_file(service_file).with_content(%r{--require-kubeconfig})
        should_not contain_file(service_file).with_content(%r{--api-servers=})
      end
    end
  end


  context 'flag --client-ca-file' do
    let(:params) { {'ca_file' => '/tmp/ca.pem' } }
    context 'versions before 1.5' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.4.8'}
        """
      ]}
      it { should_not contain_file(service_file).with_content(%r{--client-ca-file=/tmp/ca\.pem}) }
    end

    context 'versions 1.5+' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.5.0'}
        """
      ]}
      it { should contain_file(service_file).with_content(%r{--client-ca-file=/tmp/ca\.pem}) }
    end
  end


end
