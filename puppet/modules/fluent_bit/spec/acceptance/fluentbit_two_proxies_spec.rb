require 'spec_helper_acceptance'

describe '::fluent_bit' do
  pp = <<-EOS
fluent_bit::output{"test1":
  config => { 
    "elasticsearch" => {
      "host" => "search-logs-example.eu-west-1.es.amazonaws.com",
      "port" => 443,
      "tls" => true,
      "tlsVerify" => true,
      "awsESProxy" => {
        "port" => 9200
      },
    },
    "types" => ["all"],
  },
}
fluent_bit::output{"test2":
  config => { 
    "elasticsearch" => {
      "host" => "search-logs-example.eu-west-1.es.amazonaws.com",
      "port" => 443,
      "tls" => true,
      "tlsVerify" => true,
      "awsESProxy" => {
        "port" => 9201
      },
    },
    "types" => ["all"],
  },
}
EOS

  before(:all) do
    # assign private ip addresses
    hosts.each do |host|
      # clear firewall
      on host, "iptables -F INPUT"
    end
  end

  it 'should setup fluent bit without errors based on the example' do
    hosts.each do |host|
      apply_manifest_on(host, pp, :catch_failures => true)
      expect(
        apply_manifest_on(host, pp, :catch_failures => true).exit_code
      ).to be_zero
    end
  end
end