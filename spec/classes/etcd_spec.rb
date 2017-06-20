require 'spec_helper'

describe 'tarmak::etcd' do
  let(:pre_condition) {[
    """
class{'vault_client': token => 'test-token'}
"""
  ]}
  let(:facts) {{
    :hostname => 'etcd-1'
  }}

  context 'without params' do
    it do
      is_expected.to compile
    end
  end

  context '3 node etcd cluster with start index 0' do
    let(:pre_condition) {[
      """
        class{'vault_client': token => 'test-token'}
        class{'tarmak': etcd_start_index => 0}
      """
    ]}

    context 'on node etcd-0' do
      let(:facts) {{
        :hostname => 'etcd-0'
      }}
      it do
        is_expected.to compile
      end
    end

    context 'on node etcd-3' do
      let(:facts) {{
        :hostname => 'etcd-3'
      }}
      it do
        is_expected.to compile.and_raise_error(/is not within the etcd_cluster/)
      end
    end
  end

  context '3 node etcd cluster with start index 1' do
    let(:pre_condition) {[
      """
        class{'vault_client': token => 'test-token'}
        class{'tarmak':}
      """
    ]}

    context 'on node etcd-0' do
      let(:facts) {{
        :hostname => 'etcd-0'
      }}
      it do
        is_expected.to compile.and_raise_error(/is not within the etcd_cluster/)
      end
    end

    context 'on node etcd-1' do
      let(:facts) {{
        :hostname => 'etcd-1'
      }}
      it do
        is_expected.to compile
      end
    end
  end
end
