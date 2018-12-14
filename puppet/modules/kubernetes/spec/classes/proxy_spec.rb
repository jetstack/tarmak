require 'spec_helper'

describe 'kubernetes::proxy' do

  let :service_file do
      '/etc/systemd/system/kube-proxy.service'
  end

  let :proxy_config do
      '/etc/kubernetes/kube-proxy-config.yaml'
  end

  let :service_name do
    'kube-proxy.service'
  end

  context 'proxy config' do
    context 'on kubernetes 1.10' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.10.0'}
        """
      ]}
      it 'is not used' do
        should_not contain_file(service_file).with_content(%r{--config=/etc/kubernetes/kube-proxy-config\.yaml})
        should_not contain_file(proxy_config)
      end
    end

    context 'on kubernetes 1.11' do
      let(:pre_condition) {[
        """
        class{'kubernetes': version => '1.11.0'}
        """
      ]}
      it 'is used' do
          should contain_file(service_file).with_content(%r{--config=/etc/kubernetes/kube-proxy-config\.yaml})
          should contain_file(proxy_config)
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
            should_not contain_file(proxy_config).with_content(%r{featureGates:})
          end
        end

        context 'some' do
          let(:params) { {
            "feature_gates" => {"PodPriority" => true, "foobar" => false, "foo" => true, "edge=case" => true}
          }}
          it 'config contain' do
            should contain_file(proxy_config).with_content(%r{featureGates:\n})
            should contain_file(proxy_config).with_content(%r{  PodPriority: true\n})
            should contain_file(proxy_config).with_content(%r{  foobar: false\n})
            should contain_file(proxy_config).with_content(%r{  foo: true\n})
            should contain_file(proxy_config).with_content(%r{  edge=case: true})
          end
        end
      end

  context 'defaults' do
    it do
      is_expected.to compile
      should contain_service(service_name).with_ensure('running')
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
