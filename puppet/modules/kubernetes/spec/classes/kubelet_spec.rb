require 'spec_helper'

describe 'kubernetes::kubelet' do

  let :service_file do
      '/etc/systemd/system/kubelet.service'
  end

  let :kubeconfig_file do
      '/etc/kubernetes/kubeconfig-kubelet'
  end

  let :kubelet_config do
      '/var/lib/kubelet/kubelet-config.yaml'
  end

  context 'defaults' do
    it do
      should contain_file(service_file).with_content(/--register-node=true/)
      should contain_file(service_file).with_content(/--node-labels=role=worker/)
      should contain_file(service_file).with_content(/--cluster-dns=10.254.0.10/)
      should contain_file(service_file).with_content(/--cluster-domain=cluster.local/)
      should contain_file(service_file).with_content(/--allow-privileged=true/)
      should_not contain_file(service_file).with_content(/--network-plugin/)
      should contain_file(service_file).with_content(/--container-runtime=docker/)
      should contain_file(service_file).with_content(%r{--kubeconfig=/etc/kubernetes/kubeconfig-kubelet})
      should contain_file(service_file).with_content(%r{--eviction-hard=memory.available<5%})
      should contain_file(service_file).with_content(%r{--eviction-minimum-reclaim=memory.available=100Mi,nodefs.available=1Gi})
      should contain_file(service_file).with_content(%r{--eviction-soft=memory.available<10%,nodefs.available<15%,nodefs.inodesFree<10%})
      should contain_file(service_file).with_content(%r{--eviction-soft-grace-period=memory.available=0m,nodefs.available=0m,nodefs.inodesFree=0m})
      should contain_file(service_file).with_content(%r{--eviction-max-pod-grace-period=-1})
      should contain_file(service_file).with_content(%r{--eviction-pressure-transition-period=2m})
    end
  end

  context 'without soft evictions' do
    let(:params) { {
      "eviction_soft_enabled" => false
    }}
    it do
      should_not contain_file(service_file).with_content(%r{--eviction-soft})
      should_not contain_file(service_file).with_content(%r{--eviction-soft-grace-period})
      should_not contain_file(service_file).with_content(%r{--eviction-max-pod-grace-period})
      should_not contain_file(service_file).with_content(%r{--eviction-pressure-transition-period})
      should contain_file(service_file).with_content(%r{--eviction-hard=memory.available<5%})
      should contain_file(service_file).with_content(%r{--eviction-minimum-reclaim=memory.available=100Mi,nodefs.available=1Gi})
    end
  end

  context 'soft evictions with modifications' do
    let(:params) { {
      "eviction_soft_memory_available_threshold" => '15%',
      "eviction_minimum_reclaim_nodefs_available" => '2Gi',
      "eviction_soft_nodefs_inodes_free_grace_period" => '1m',
      "eviction_max_pod_grace_period" => '300',
      "eviction_pressure_transition_period" => '5m'
    }}
    it do
      should contain_file(service_file).with_content(%r{--eviction-soft=memory.available<15%,nodefs.available<15%,nodefs.inodesFree<10%})
      should contain_file(service_file).with_content(%r{--eviction-minimum-reclaim=memory.available=100Mi,nodefs.available=2Gi})
      should contain_file(service_file).with_content(%r{--eviction-soft-grace-period=memory.available=0m,nodefs.available=0m,nodefs.inodesFree=1m})
      should contain_file(service_file).with_content(%r{--eviction-max-pod-grace-period=300})
      should contain_file(service_file).with_content(%r{--eviction-pressure-transition-period=5m})
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
        },
        'os' => {
          'selinux' => {
            'enabled' => true
          }
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
      should contain_file(parent_dir).with_seltype('container_file_t')
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
    let(:params) { { 'role' => 'master' } }

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

      context 'with additional node taints' do
        let(:params) { { 'role' => 'master', 'node_taints' => { 'foo' => 'bar:NoSchedule' } } }

        it "retains the default taints" do
          have_service_file = contain_file('/etc/systemd/system/kubelet.service')
          should have_service_file.with_content(/--register-with-taints=node-role\.kubernetes\.io\/master=:NoSchedule,foo=bar:NoSchedule/)
        end
      end

      context 'with overriding node taints' do
        let(:params) { { 'role' => 'master', 'node_taints' => { 'node-role.kubernetes.io/master' => ':NoExecute' } } }

        it "replaces the default" do
          have_service_file = contain_file('/etc/systemd/system/kubelet.service')
          should have_service_file.with_content(/--register-with-taints=node-role.kubernetes.io\/master=:NoExecute/)
        end
      end

      context 'with blank overriding node taints' do
        let(:params) { { 'role' => 'master', 'node_taints' => { 'node-role.kubernetes.io/master' => 'REMOVE:REMOVE' } } }

        it "removes the taint" do
          have_service_file = contain_file('/etc/systemd/system/kubelet.service')
          should_not have_service_file.with_content(/--register-with-taints/)
        end
      end
    end
  end

  context 'with role worker' do
    let(:params) { { 'role' => 'worker' } }

    it do
      have_service_file = contain_file('/etc/systemd/system/kubelet.service')
      should have_service_file.with_content(/--node-labels=role=worker,node-role.kubernetes.io\/worker="/)
    end

    context 'with additional node labels' do
      let(:params) { { 'role' => 'worker', 'node_labels' => { 'foo' => 'bar' } } }

      it do
        have_service_file = contain_file('/etc/systemd/system/kubelet.service')
        should have_service_file.with_content(/--node-labels=role=worker,node-role.kubernetes.io\/worker=,foo=bar/)
      end
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
    let(:params) {{
      'client_ca_file' => '/tmp/client_ca.pem'
    }}
    context 'versions before 1.5' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.4.8'}
        """
      ]}
      it { should_not contain_file(service_file).with_content(%r{--client-ca-file=/tmp/client_ca\.pem}) }
    end

    context 'versions 1.5+' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.5.0'}
        """
      ]}
      it {
        should contain_file(service_file).with_content(%r{--client-ca-file=/tmp/client_ca\.pem})
      }
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
        it { should contain_file(service_file).with_content(%r{--cgroup-driver=cgroupfs}) }
      end

      context 'on anything but redhat family os' do
        let(:facts) { {'osfamily' => 'Debian' } }
        it { should contain_file(service_file).with_content(%r{--cgroup-driver=cgroupfs}) }
      end
    end
  end

  ['kube', 'system'].each do |cgroup_type|
    context 'runtime cgroups reserved' do
      let(:facts) { {'kernelversion' => '4.14.1' } }
      let(:pre_condition) {[
        """
          class{'kubernetes': version => '1.9.5'}
        """
      ]}

      context 'with both cpu and memory a supplied' do
        let(:params) { {
          "cgroup_#{cgroup_type}_reserved_cpu"    => '100m',
          "cgroup_#{cgroup_type}_reserved_memory" => '128Mi',
        }}
        it do
          should contain_file(service_file).with_content(%r{--#{cgroup_type}-reserved=cpu=100m,memory=128Mi})
        end
      end

      context 'with only cpu supplied' do
        let(:params) { {
          "cgroup_#{cgroup_type}_reserved_cpu"    => '100m',
          "cgroup_#{cgroup_type}_reserved_memory" => nil,
        }}
        it do
          should contain_file(service_file).with_content(%r{--#{cgroup_type}-reserved=cpu=100m})
        end
      end

      context 'with only memory supplied' do
        let(:params) { {
          "cgroup_#{cgroup_type}_reserved_cpu"    => nil,
          "cgroup_#{cgroup_type}_reserved_memory" => '128Mi',
        }}
        it do
          should contain_file(service_file).with_content(%r{--#{cgroup_type}-reserved=memory=128Mi})
        end
      end

      context 'with nothing supplied' do
        let(:params) { {
          "cgroup_#{cgroup_type}_reserved_cpu"    => nil,
          "cgroup_#{cgroup_type}_reserved_memory" => nil,
        }}
        it do
          should_not contain_file(service_file).with_content(%r{--#{cgroup_type}-reserved=})
        end
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

  context 'kubelet config' do
    context 'on kubernetes 1.10' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.10.0'}
        """
      ]}
      it 'is not used' do
        should_not contain_file(service_file).with_content(%r{--config=/var/lib/kubelet/kubelet-config\.yaml})
        should_not contain_file(kubelet_config)
      end
    end

    context 'on kubernetes 1.11' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.11.0'}
        """
      ]}
      let(:params) { {
        "feature_gates" => ["PodPriority=true", "foobar=true", "foo", "edge=case=true"]
      }}
      it 'is used' do
          should contain_file(service_file).with_content(%r{--config=/var/lib/kubelet/kubelet-config\.yaml})
          should contain_file(kubelet_config).with_content(%r{PodPriority: true})
          should contain_file(kubelet_config).with_content(%r{foobar: true})
          should contain_file(kubelet_config).with_content(%r{foo: true})
          should contain_file(kubelet_config).with_content(%r{edge=case: true})
      end
    end

  end
end
