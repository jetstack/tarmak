require 'spec_helper_acceptance'

module Helpers
  def log_magic_line
    hosts.each do |host|
      on host, "logger \"#{@magic_logline}\""
    end
  end

  def find_magic_line
    hosts.each do |host|
      result = JSON.parse(host.execute("curl --retry 5 --fail -s -X GET 'http://localhost:9200/_all/_search?q=MESSAGE:\"#{@magic_logline}\"'"))
      return nil unless result.key? 'hits'
      return nil unless result['hits'].key? 'total'
      total = result['hits']['total']
      fail 'more than a single result found' if total > 1
      return nil if total < 1
      return result['hits']['hits'][0]
    end
  end

  def retry_find_magic_line(tries=100)
    count = 0
    while true do
      log_line = find_magic_line
      return log_line unless log_line.nil?

      count += 1
      fail "could not find magic logline after #{count} tries" if count > tries
      sleep 1
    end
    return nil
  end

  def document_count
    count = 0
    result = JSON.parse(host.execute("curl --retry 5 --fail -s -X GET 'http://localhost:9200/_cat/indices?format=json"))
    $stderr.puts "indices = #{result}"
    # TODO: Count here
    count
  end
end

RSpec.configure do |c|
  c.include Helpers
end

describe '::fluent_bit' do
  before(:all) do
    hosts.each do |host|
      # clear firewall
      on host, 'iptables -F INPUT'

      # make sure golang uses proper hosts file
      on host, 'echo "hosts: files dns" > /etc/nsswitch.conf'

      # create some fake aws creds
      on host, 'echo -e "AWS_ACCESS_KEY_ID=AKID1234567890\nAWS_SECRET_ACCESS_KEY=MY-SECRET-KEY" > /etc/sysconfig/aws-es-proxy-test'

      # setup a local elasticsearch instance
      on host, "cp #{$module_path}fluent_bit/spec/fixtures/elasticsearch.repo /etc/yum.repos.d"
      on host, 'rpm --import https://artifacts.elastic.co/GPG-KEY-elasticsearch'
      on host, 'yum -y install java-1.8.0-openjdk elasticsearch wget'
      on host, 'systemctl start elasticsearch.service'
      on host, 'wget --retry-connrefused -T 60 http://localhost:9200/_cluster/health -O /dev/null 2> /dev/null'
    end
  end

  let(:port) do
    9200
  end

  context 'fluentbit with aws-es proxy' do
    before(:all) do
      @magic_logline = "magic-logline-#{SecureRandom.uuid}"
      @magic_prefix = "magic-prefix-#{SecureRandom.uuid}"
    end

    let(:pp) do
      pp = <<-EOS
host { 'fake.eu-west-1.es.amazonaws.com':
 ensure => present,
 ip => '127.0.0.1',
}
-> fluent_bit::output{"test":
    config => {
        "elasticsearch" => {
            "host" => "fake.eu-west-1.es.amazonaws.com",
            "port" => #{port},
            "tls" => false,
            "amazonESProxy" => {
              "port" => 9201
            },
          },
        "types" => ["all"],
    },
}
      EOS
      pp
    end


    it 'should setup fluent bit without errors in first apply' do
      hosts.each do |host|
        apply_manifest_on(host, pp, catch_failures: true)
        expect(
          apply_manifest_on(host, pp, catch_failures: true).exit_code
        ).to be_zero
      end
    end

    it 'should receive magic log line' do
      log_magic_line
      log_line = retry_find_magic_line
      expect(log_line).not_to be_nil
      expect(log_line['_index']).to start_with('logstash-')
    end
  end

  context 'fluentbit with aws-es proxy using custom index' do
    before(:all) do
      @magic_logline = "magic-logline-#{SecureRandom.uuid}"
      @magic_prefix = "magic-prefix-#{SecureRandom.uuid}"
    end

    let(:pp) do
      pp = <<-EOS
host { 'fake.eu-west-1.es.amazonaws.com':
 ensure => present,
 ip => '127.0.0.1',
}
-> fluent_bit::output{"test":
    config => {
        "elasticsearch" => {
            "host" => "fake.eu-west-1.es.amazonaws.com",
            "port" => #{port},
            "tls" => false,
            "logstashPrefix" => "#{@magic_prefix}",
            "amazonESProxy" => {
              "port" => 9201
            },
          },
        "types" => ["all"],
    },
}
      EOS
      pp
    end

    it 'should setup fluent bit without errors in first apply' do
      hosts.each do |host|
        apply_manifest_on(host, pp, catch_failures: true)
        expect(
          apply_manifest_on(host, pp, catch_failures: true).exit_code
        ).to be_zero
      end
    end

    it 'should receive magic log line and prefix' do
      log_magic_line
      log_line = retry_find_magic_line
      expect(log_line).not_to be_nil
      expect(log_line['_index']).to start_with(@magic_prefix)
    end
  end

  context 'fluentbit without any output' do
    before(:all) do
      @magic_logline = "magic-logline-#{SecureRandom.uuid}"
      @magic_prefix = "magic-prefix-#{SecureRandom.uuid}"
    end

    let(:pp) do
      pp = <<-EOS
host { 'fake.eu-west-1.es.amazonaws.com':
 ensure => present,
 ip => '127.0.0.1',
}
-> fluent_bit::output{"test":
    config => {
        "types" => ["all"],
    },
}
      EOS
      pp
    end

    it 'should setup fluent bit without errors in first apply' do
      hosts.each do |host|
        apply_manifest_on(host, pp, catch_failures: true)
        expect(
          apply_manifest_on(host, pp, catch_failures: true).exit_code
        ).to be_zero
      end
    end

    it 'should not receive magic log line and prefix' do
      log_magic_line

      expect{retry_find_magic_line(tries=10)}.to raise_error(/could not find magic logline/)
    end
  end
end
