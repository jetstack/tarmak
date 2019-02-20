require 'spec_helper_acceptance'
require 'webrick'

class ESMock
  attr_reader :port, :magic_logline, :magic_prefix, :semaphore
  attr_accessor :magic_logline_appeared, :magic_prefix_appeared

  def initialize
    reset
    @semaphore = Mutex.new
    @port = 9200
    @http_server = WEBrick::HTTPServer.new(Port: @port)
    @http_server.mount '/', ESMockHandle, self
  end

  def start
    # listen to local ES port
    trap('INT') { @http_server.shutdown }
    @http_server_thread = Thread.start { @http_server.start }
  end

  def stop
    @http_server.shutdown
    @http_server_thread.join
  end

  def reset
    @magic_logline = "magic-logline-#{SecureRandom.uuid}"
    @magic_prefix = "magic-prefix-#{SecureRandom.uuid}"
    @magic_logline_appeared = nil
    @magic_prefix_appeared = nil
  end
end

class ESMockHandle < WEBrick::HTTPServlet::AbstractServlet
  @server = nil

  def initialize(i, server)
    super i
    @server = server
  end

  def do_POST(request, response)
    key_message = 'MESSAGE'
    key_index = 'index'
    key__index = '_index'

    index = nil
    count = 0
    request.body.each_line do |line|
      log_line = JSON.parse(line)

      if log_line.key?(key_index) and log_line[key_index].key?(key__index)
        index = log_line[key_index][key__index]
        if index.include? @server.magic_prefix
          $stderr.puts "magic prefix appeared: '#{log_line}'"
          @server.semaphore.synchronize do
            @server.magic_prefix_appeared = true
          end
        end
        next
      end

      count += 1

      if log_line.key?(key_message) and log_line[key_message].include?(@server.magic_logline)
        $stderr.puts "magic logline appeared: '#{log_line}'"
        @server.semaphore.synchronize do
          @server.magic_logline_appeared = true
        end
      end
    end

    $stderr.puts "received #{count} log lines for index '#{index}'"

    response.status = 200
    response.body = '{"took":30,"errors":false,"items":[]}'
  end
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
    end

    @mock_es = ESMock.new
    @mock_es.start
  end


  after(:all) do
    @mock_es.stop
  end

  context 'fluentbit with aws-es proxy' do
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
            "port" => #{@mock_es.port},
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

    before(:all) do
      @mock_es.reset
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
      hosts.each do |host|
        on host, "logger \"#{@mock_es.magic_logline}\""
      end

      count = 0
      while true
        count += 1
        fail "no log line received after 1000 tries" if count > 1000
        break unless @mock_es.magic_logline_appeared.nil?
        sleep(1)
      end

      expect(@mock_es.magic_logline_appeared).to equal(true)
      expect(@mock_es.magic_prefix_appeared).to equal(nil)
    end
  end

  context 'fluentbit with aws-es proxy using custom index' do
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
             "port" => #{@mock_es.port},
             "tls" => false,
             "logstashPrefix" => "#{@mock_es.magic_prefix}",
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

    before(:all) do
      @mock_es.reset
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
      hosts.each do |host|
        on host, "logger \"#{@mock_es.magic_logline}\""
      end

      count = 0
      while true
        count += 1
        fail "no log line received after 1000 tries" if count > 1000
        break unless @mock_es.magic_logline_appeared.nil? or @mock_es.magic_logline_appeared.nil?
        sleep(1)
      end

      expect(@mock_es.magic_logline_appeared).to equal(true)
      expect(@mock_es.magic_prefix_appeared).to equal(true)
    end
  end

  context 'fluentbit without any output' do
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

    before(:all) do
      @mock_es.reset
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
      hosts.each do |host|
        on host, "logger \"#{@mock_es.magic_logline}\""
      end

      sleep(10)

      expect(@mock_es.magic_logline_appeared).to equal(nil)
      expect(@mock_es.magic_prefix_appeared).to equal(nil)
    end
  end
end
