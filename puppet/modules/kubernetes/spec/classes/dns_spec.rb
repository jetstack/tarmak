require 'spec_helper'

describe 'kubernetes::dns' do
  let :kube_service_file_apply do
    '/etc/systemd/system/kubectl-apply-kube-dns.service'
  end

  let :kube_service_file_delete do
    '/etc/systemd/system/kubectl-delete-kube-dns.service'
  end

  let :core_service_file_apply do
    '/etc/systemd/system/kubectl-apply-core-dns.service'
  end

  let :core_service_file_delete do
    '/etc/systemd/system/kubectl-delete-core-dns.service'
  end

  let :kube_apply_manifests_file do
    '/etc/kubernetes/apply/kube-dns.yaml'
  end

  let :core_apply_manifests_file do
    '/etc/kubernetes/apply/core-dns.yaml'
  end

  let :kube_delete_manifests_file do
    '/etc/kubernetes/delete/kube-dns.yaml'
  end

  let :core_delete_manifests_file do
    '/etc/kubernetes/delete/core-dns.yaml'
  end

  context 'with default values for all parameters' do
    let(:pre_condition) {[
        """
      include kubernetes::apiserver,
      class{'kubernetes': version => '1.9.0'}
        """
    ]}

    it { should contain_class('kubernetes::dns') }
    it { should contain_class('kubernetes::apiserver') }

    it 'should write systemd unit for applying' do
      should contain_file(kube_service_file_apply)
      should contain_file(core_service_file_delete)
    end

    it 'should write manifests' do
      should contain_file(kube_apply_manifests_file).with_content(/--domain=cluster\.local\./)
      should contain_file(kube_apply_manifests_file).with_content(/clusterIP: 10\.254\.0\.10/)
      should contain_file(core_delete_manifests_file)
    end
  end

  context 'with version 1.11' do
    let(:pre_condition) {[
        """
      include kubernetes::apiserver,
      class{'kubernetes': version => '1.11.0'}
        """
    ]}

    it { should contain_class('kubernetes::dns') }
    it { should contain_class('kubernetes::apiserver') }

    it 'should write systemd unit for applying' do
      should contain_file(core_service_file_apply).with_content(/User=kubernetes/)
      should contain_file(kube_service_file_delete).with_content(/User=kubernetes/)
    end

    it 'should write manifests' do
      should contain_file(core_apply_manifests_file)
      should contain_file(kube_delete_manifests_file).with_content(/--domain=cluster\.local\./)
      should contain_file(kube_delete_manifests_file).with_content(/clusterIP: 10\.254\.0\.10/)
    end
  end
end
