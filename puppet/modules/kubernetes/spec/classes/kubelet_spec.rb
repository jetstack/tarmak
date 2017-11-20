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
      should contain_file(service_file).with_content(/--allow-privileged=true/)
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

  context 'aws ebs with SElinux' do
    let(:facts) do
      {
        'ec2_metadata' => {
          'placement' => {
            'availability-zone' => 'my-zone-1z',
          },
        }
      }
    end

    let(:pre_condition) {[
      """
        class{'kubernetes': cloud_provider => 'aws'}
      """
    ]}

    let(:parent_dir) do
      '/var/lib/kubelet/plugins/kubernetes.io/aws-ebs/mounts/aws/my-zone-1z'
    end

    it 'should have seltype set on parent dir' do
      should contain_file(parent_dir).with_ensure('directory')
      should contain_file(parent_dir).with_seltype('svirt_sandbox_file_t')
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
    context 'versions before 1.6' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.5.8'}
        """
      ]}
      it do
        have_service_file = contain_file('/etc/systemd/system/kubelet.service')
        should have_service_file.with_content(/--register-schedulable=false/)
        should have_service_file.with_content(/--node-labels=role=master/)
      end
    end

    context 'versions 1.6+' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.6.0'}
        """
      ]}
      it do
        have_service_file = contain_file('/etc/systemd/system/kubelet.service')
        should have_service_file.with_content(/--register-with-taints=node-role\.kubernetes\.io\/master=:NoSchedule/)
        should have_service_file.with_content(/--node-labels=role=master/)
      end
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

  context 'flag --cgroup-driver' do
    context 'versions before 1.6' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.5.4'}
        """
      ]}
      it { should_not contain_file(service_file).with_content(%r{--cgroup-driver}) }
    end

    context 'versions 1.6+' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.6.0'}
        """
      ]}

      context 'on redhat family os' do
        let(:facts) { {'osfamily' => 'RedHat' } }
        it { should contain_file(service_file).with_content(%r{--cgroup-driver=systemd}) }
        it { should contain_file(service_file).with_content(%r{--runtime-cgroups=/systemd/system.slice}) }
        it { should contain_file(service_file).with_content(%r{--kubelet-cgroups=/systemd/system.slice}) }
      end

      context 'on anything but redhat family os' do
        let(:facts) { {'osfamily' => 'Debian' } }
        it { should contain_file(service_file).with_content(%r{--cgroup-driver=cgroupfs}) }
      end
    end
  end

  context 'kernel 4.9+ cgropus hotfix' do
    let(:facts) { {'kernelversion' => '4.10.1' } }
    let(:pre_condition) {[
      """
        class{'kubernetes': version => '1.6.11'}
      """
    ]}
    context 'on kubernetes 1.6 and newer kernels' do
      it 'is enabled' do
        should contain_file(service_file).with_content(%r{--cgroups-per-qos=false})
        should contain_file(service_file).with_content(%r{--enforce-node-allocatable= })
      end
    end
    context 'on kubernetes 1.6 and older kernels' do
      let(:facts) { {'kernelversion' => '4.1.1' } }
      it 'is disabled' do
        should_not contain_file(service_file).with_content(%r{--cgroups-per-qos=false})
        should_not contain_file(service_file).with_content(%r{--enforce-node-allocatable= })
      end
    end
    context 'on kubernetes 1.5' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.5.8'}
        """
      ]}
      it 'is disabled' do
        should_not contain_file(service_file).with_content(%r{--cgroups-per-qos=false})
        should_not contain_file(service_file).with_content(%r{--enforce-node-allocatable= })
      end
    end
  end
end
