require 'spec_helper'

describe 'vault_client::cert_service', :type => :define do

  let(:title) do
    'test1'
  end

  let(:service_name) do
    "#{title}-cert.service"
  end

  let(:service_trigger) do
    "#{title}-cert-trigger"
  end

  let(:service_trigger_command) do
    "systemctl start #{service_name}"
  end

  let(:timer_name) do
    "#{title}-cert.timer"
  end

  let(:service_file) do
    "/etc/systemd/system/#{service_name}"
  end
  
  let(:timer_file) do
    "/etc/systemd/system/#{timer_name}"
  end

  let(:pre_condition) {[
    "
class{'vault_client':
  token => 'token1'
}
    "
  ]}

  let(:params) do
    {
      :common_name => 'commonname1',
      :role => 'role1',
      :base_path => '/tmp/test',
    }
  end

  context 'should create a vault cert service' do
    it do
      should contain_service(timer_name)
      should contain_file(service_file).with_content(/EnvironmentFile=\/etc\/vault\/config/)
      should contain_file(timer_file)
      should contain_exec(service_trigger).with_command(service_trigger_command)
    end
  end

  context 'with run_exec => false' do
    let(:params) do
      super().merge({ 'run_exec' => false })
    end

    it do
      should contain_service(timer_name)
      should contain_file(service_file).with_content(/EnvironmentFile=\/etc\/vault\/config/)
      should contain_file(timer_file)
      should contain_exec(service_trigger).with_command('/bin/true')
    end
  end
end
