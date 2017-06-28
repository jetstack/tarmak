require 'spec_helper'
describe 'calico' do
  let(:pre_condition) do
    "
      class kubernetes{}
      define kubernetes::apply(
        $manifests,
      ){}
    "
  end

  context 'on cloud_provider aws' do
    let(:params) {
      {
        :cloud_provider => 'aws'
      }
    }

    it do
      should contain_class('calico')
      should contain_class('calico::disable_source_destination_check')
      should contain_file('/opt/bin/disable-source-destination-check.sh').with({
        'ensure' => 'file',
        'mode'   => '0755',
      })
    end
  end
end
