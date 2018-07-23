require 'spec_helper'

describe 'kubernetes::kubelet' do

  let :service_file do
      '/etc/systemd/system/kubelet.service'
  end

  let :kubelet_config do
      '/etc/kubernetes/kubelet-config.yaml'
  end
  
  let :service_name do
    'kubelet.service'
  end

  let :kubeconfig_file do
      '/etc/kubernetes/kubeconfig-kubelet'
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
      should contain_service(service_name).with_ensure('running')
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

  context 'feature gates' do
    context 'none' do
      let(:pre_condition) {[
          """
          class{'kubernetes': enable_pod_priority => false}
          """
      ]}
      let(:params) { {
        "feature_gates" => {}
      }}
      it 'none with no pod priority' do
        should_not contain_file(service_file).with_content(%r{--feature-gates=})
      end
    end

    context 'none + pod priority' do
      let(:pre_condition) {[
        """
          class{'kubernetes': enable_pod_priority => true}
        """
      ]}
      it 'none with pod priority' do
        should contain_file(service_file).with_content(%r{--feature-gates=PodPriority=true})
      end
    end

    context 'some' do
      let(:params) { {
        "feature_gates" => {"PodPriority" => true, "foobar" => false, "foo" => true, "edge=case" => true}
      }}
      it do
        should contain_file(service_file).with_content(%r{--feature-gates=PodPriority=true,foobar=false,foo=true,edge=case=true \\\n})
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

    # using kubelet config file
    context 'on kubernetes 1.11' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.11.0'}
        """
      ]}

      it do
        should contain_file(service_file).with_content(%r{--config=/etc/kubernetes/kubelet-config\.yaml})
      end

      context 'defaults' do
        it 'service not contain' do
          should_not contain_file(service_file).with_content(%r{--cluster-dns=})
          should_not contain_file(service_file).with_content(%r{--cluster-domain=})
          should_not contain_file(service_file).with_content(%r{--pod-cidr=})
        end

        it 'config contain' do
          should contain_file(kubelet_config).with_content(/kind: KubeletConfiguration/)
          should contain_file(kubelet_config).with_content(/apiVersion: kubelet.config.k8s.io\/v1beta1/)
          should contain_file(kubelet_config).with_content(%r{clusterDNS:\n  - 10.254.0.10})
          should contain_file(kubelet_config).with_content(%r{clusterDomain: cluster.local})
        end
      end

      context 'auth' do
        let(:params) {{
          'client_ca_file' => '/tmp/client_ca.pem'
        }}

        it 'service not contain' do
          should_not contain_file(service_file).with_content(%r{--client-ca-file=/tmp/client_ca\.pem})
        end

        it 'config contain' do
          should contain_file(kubelet_config).with_content(%r{authentication:})
          should contain_file(kubelet_config).with_content(%r{  x509:})
          should contain_file(kubelet_config).with_content(%r{    clientCAFile: /tmp/client_ca\.pem})
          should contain_file(kubelet_config).with_content(%r{  anonymous:})
          should contain_file(kubelet_config).with_content(%r{    enabled: false})
          should contain_file(kubelet_config).with_content(%r{  webook:})
          should contain_file(kubelet_config).with_content(%r{    enabled: true})
          should contain_file(kubelet_config).with_content(%r{authorization:})
          should contain_file(kubelet_config).with_content(%r{  mode: Webhook})
        end
      end

      context 'feature gates' do
        context 'none' do
          let(:pre_condition) {[
              """
              class{'kubernetes': enable_pod_priority => false}
              """
          ]}
          let(:params) { {
            "feature_gates" => {}
          }}
          it 'none with no pod priority' do
            should_not contain_file(kubelet_config).with_content(%r{featureGates:})
          end
        end

        context 'none + pod priority' do
          let(:pre_condition) {[
            """
              class{'kubernetes': enable_pod_priority => true, version => '1.11.1'}
            """
          ]}
          it 'none with pod priority' do
            should contain_file(kubelet_config).with_content(%r{featureGates:\n  PodPriority: true})
          end
        end

        context 'some' do
          let(:params) { {
            "feature_gates" => {"PodPriority" => true, "foobar" => false, "foo" => true, "edge=case" => true}
          }}
          it 'config contain' do
            should contain_file(kubelet_config).with_content(%r{featureGates:\n})
            should contain_file(kubelet_config).with_content(%r{  PodPriority: true\n})
            should contain_file(kubelet_config).with_content(%r{  foobar: false\n})
            should contain_file(kubelet_config).with_content(%r{  foo: true\n})
            should contain_file(kubelet_config).with_content(%r{  edge=case: true\n})
          end
        end
      end

      context 'cgroups' do
        it do
          should_not contain_file(service_file).with_content(%r{--cgroup-driver=})
          should_not contain_file(service_file).with_content(%r{--cgroup-root=})

          should contain_file(kubelet_config).with_content(%r{cgroupDriver: cgroupfs})
          should contain_file(kubelet_config).with_content(%r{cgroupRoot: /})
          should contain_file(kubelet_config).with_content(%r{kubeletCgroups: /podruntime.slice})
          should contain_file(kubelet_config).with_content(%r{systemCgroups: /system.slice})
        end

        ['kube', 'system'].each do |cgroup_type|
          context 'runtime cgroups reserved' do

            context 'with both cpu and memory a supplied' do
              let(:params) { {
                "cgroup_#{cgroup_type}_reserved_cpu"    => '100m',
                "cgroup_#{cgroup_type}_reserved_memory" => '128Mi',
              }}
              it do
                should_not contain_file(service_file).with_content(%r{--#{cgroup_type}-reserved=cpu=100m,memory=128Mi})
                should contain_file(kubelet_config).with_content(%r{#{cgroup_type}Reserved:\n  cpu: 100m\n  memory: 128Mi})
              end
            end

            context 'with only cpu supplied' do
              let(:params) { {
                "cgroup_#{cgroup_type}_reserved_cpu"    => '100m',
                "cgroup_#{cgroup_type}_reserved_memory" => nil,
              }}
              it do
                should_not contain_file(service_file).with_content(%r{--#{cgroup_type}-reserved=cpu=100m})
                should contain_file(kubelet_config).with_content(%r{#{cgroup_type}Reserved:\n  cpu: 100m})
                should_not contain_file(kubelet_config).with_content(%r{#{cgroup_type}Reserved:\n  cpu: 100m\n  memory:})
              end
            end

            context 'with only memory supplied' do
              let(:params) { {
                "cgroup_#{cgroup_type}_reserved_cpu"    => nil,
                "cgroup_#{cgroup_type}_reserved_memory" => '128Mi',
              }}
              it do
                should_not contain_file(service_file).with_content(%r{--#{cgroup_type}-reserved=memory=128Mi})
                should contain_file(kubelet_config).with_content(%r{#{cgroup_type}Reserved:\n  memory: 128Mi})
                should_not contain_file(kubelet_config).with_content(%r{#{cgroup_type}Reserved:\n  cpu:})
              end
            end

            context 'with nothing supplied' do
              let(:params) { {
                "cgroup_#{cgroup_type}_reserved_cpu"    => nil,
                "cgroup_#{cgroup_type}_reserved_memory" => nil,
              }}
              it do
                should_not contain_file(service_file).with_content(%r{--#{cgroup_type}-reserved=})
                should contain_file(kubelet_config).with_content(%r{#{cgroup_type}Reserved:})
              end
            end
          end
        end
      end

      context 'tls' do
        let(:params) { {
          'cert_file' => '/etc/kubernetes/kubelet-cert.pem',
          'key_file'  => '/etc/kubernetes/kubelet-key.pem',
        }}
        context 'pre 1.11' do
          let(:pre_condition) {[
            """
            class{'kubernetes': version => '1.10.0'}
            """
          ]}
          it do
            should contain_file(service_file).with_content(%r{--tls-cert-file=/etc/kubernetes/kubelet-cert.pem})
            should contain_file(service_file).with_content(%r{--tls-private-key-file=/etc/kubernetes/kubelet-key.pem})
          end
        end

        context 'post 1.11' do
          let(:pre_condition) {[
            """
            class{'kubernetes': version => '1.11.0'}
            """
          ]}
          it do
            should_not contain_file(service_file).with_content(%r{--tls-cert-file=/etc/kubernetes/kubelet-cert.pem})
            should_not contain_file(service_file).with_content(%r{--tls-private-key-file=/etc/kubernetes/kubelet-key.pem})
            should contain_file(kubelet_config).with_content(%r{tlsCertFile: /etc/kubernetes/kubelet-cert.pem})
            should contain_file(kubelet_config).with_content(%r{tlsPrivateKeyFile: /etc/kubernetes/kubelet-key.pem})
          end
        end
      end

      context 'evictions' do
        context 'hard' do
          it do
            should_not contain_file(service_file).with_content(%r{--eviction-hard=})
            should contain_file(kubelet_config).with_content(%r{evictionHard:\n})
            should contain_file(kubelet_config).with_content(%r{  memory.available: 5%\n})
            should contain_file(kubelet_config).with_content(%r{  nodefs.available: 10%\n})
            should contain_file(kubelet_config).with_content(%r{  nodefs.inodesFree: 5%\n})
          end
        end

        context 'soft' do
          it do
            should_not contain_file(service_file).with_content(%r{--eviction-soft=})
            should contain_file(kubelet_config).with_content(%r{evictionSoft:\n})
            should contain_file(kubelet_config).with_content(%r{  memory.available: 10%\n})
            should contain_file(kubelet_config).with_content(%r{  nodefs.available: 15%\n})
            should contain_file(kubelet_config).with_content(%r{  nodefs.inodesFree: 15%\n})
          end
        end

        context 'soft grace period' do
          it do
            should_not contain_file(service_file).with_content(%r{--eviction-soft-grace-period=})
            should contain_file(kubelet_config).with_content(%r{evictionSoftGracePeriod:\n})
            should contain_file(kubelet_config).with_content(%r{  memory.available: 0m\n})
            should contain_file(kubelet_config).with_content(%r{  nodefs.available: 0m\n})
            should contain_file(kubelet_config).with_content(%r{  nodefs.inodesFree: 0m})
          end
        end

        context "minimum reclaim" do
          it do
            should_not contain_file(service_file).with_content(%r{--eviction-minimum-reclaim=})
            should contain_file(kubelet_config).with_content(%r{evictionMinimumReclaim:\n  memory.available: 100Mi\n  nodefs.available: 1Gi\n})
          end
        end

        it do
          should_not contain_file(service_file).with_content(%r{--eviction-max-pod-grace-period=})
          should_not contain_file(service_file).with_content(%r{--eviction-pressure-transition-period=})
          should contain_file(kubelet_config).with_content(%r{evictionMaxPodGracePeriod: -1})
          should contain_file(kubelet_config).with_content(%r{evictionPressureTransitionPeriod: 2m})
        end
      end
    end
  end

  context 'feature gates' do
    context 'without given feature gates and not enabled pod priority' do
      let(:params) { {'feature_gates' => {}}}
      it 'should have default feature gates' do
        should_not contain_file(service_file).with_content(/#{Regexp.escape('--feature-gates=')}/)
      end
    end

    context 'without given feature gates and enabled pod priority' do
      let(:pre_condition) {[
        """
        class{'kubernetes': enable_pod_priority => true}
        """
      ]}
      let(:version) { '1.6.0' }
      it 'should have default feature gates' do
        should contain_file(service_file).with_content(/#{Regexp.escape('--feature-gates=PodPriority=true')}/)
      end
    end

    context 'with given feature gates' do
      let(:params) { {'feature_gates' => {'foo' => true, 'bar' => true}}}
      it 'should have custom feature gates' do
        should contain_file(service_file).with_content(/#{Regexp.escape('--feature-gates=foo=true,bar=true')}/)
      end
    end
  end
  context 'with service_ensure => stopped' do
    let(:params) { { 
      "service_ensure" => 'stopped',
    }}

    it do
      should contain_service(service_name).with_ensure('stopped')
    end
  end
end
