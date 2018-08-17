require 'spec_helper'

describe 'get_ipaddress_in_network' do
  it { is_expected.not_to eq(nil) }
  it { is_expected.to run.with_params().and_raise_error(Puppet::ParseError, /wrong number of arguments/i) }
  it { is_expected.to run.with_params("one", "two").and_raise_error(Puppet::ParseError, /wrong number of arguments/i) }

  context "On Linux Systems" do
    let(:facts) do
      {
        :interfaces => 'eth0,eth1,lo',
        :ipaddress => '10.0.0.1',
        :ipaddress_lo => '127.0.0.1',
        :ipaddress_eth0 => '10.0.0.1',
        :ipaddress_eth1 => '123.0.0.1',
      }
    end

    it { is_expected.to run.with_params('127.0.0.1/32').and_return('127.0.0.1') }
    it { is_expected.to run.with_params('123.0.0.0/24').and_return('123.0.0.1') }
    it { is_expected.to run.with_params('8.8.8.0/24').and_return(:undefined) }
    it { is_expected.to run.with_params('256.8.8.0/24').and_raise_error(Puppet::ParseError, /invalid network:/i) }
  end
end
