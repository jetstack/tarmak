require 'spec_helper'

describe 'kubernetes::dns' do
  let(:pre_condition) do
    [
      'include kubernetes::apiserver'
    ]
  end

  let :service_file do
    '/etc/systemd/system/kubectl-apply-kube-dns.service'
  end

  let :manifests_file do
    '/etc/kubernetes/apply/kube-dns.yaml'
  end

  context 'with default values for all parameters' do

    it { should contain_class('kubernetes::dns') }

    it 'should write systemd unit for applying' do
      should contain_file(service_file).with_content(/User=kubernetes/)
    end

    it 'should write manifests' do
      should contain_file(manifests_file).with_content(/--domain=cluster\.local\./)
      should contain_file(manifests_file).with_content(/clusterIP: 10\.254\.0\.10/)
    end
  end
end
