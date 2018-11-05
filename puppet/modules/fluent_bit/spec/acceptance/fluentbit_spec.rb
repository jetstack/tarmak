require 'spec_helper_acceptance'
require 'webrick'

$mock_es_port = 9200
$mock_es = nil

$magic_logline = 'iamalinethatwouldneverappearinanormallog'
$magic_prefix = 'iamaprefixthatwouldneverbeprefixedtonormallogs'
$magic_logline_appeared = false
$magic_prefix_appeared = false

class ESMock < WEBrick::HTTPServlet::AbstractServlet
  def do_POST(request, response)
    if request.body.include? $magic_logline
      $magic_logline_appeared = true
    end

    if request.body.include? $magic_prefix
      $magic_prefix_appeared = true
    end

    response.status = 200
    response.body = 'ok, mock'
  end
end

describe '::fluent_bit' do
  let(:pp) do
    pp = <<-EOS
host { 'fake.eu-west-1.es.amazonaws.com':
  ensure => present,
  ip => $facts['networking']['dhcp'],
}
-> fluent_bit::output{"test":
    config => {
        "elasticsearch" => {
            "host" => "fake.eu-west-1.es.amazonaws.com",
            "port" => #{$mock_es_port},
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

   let(:pp2) do
     pp2 = <<-EOS
 host { 'fake.eu-west-1.es.amazonaws.com':
   ensure => present,
   ip => $facts['networking']['dhcp'],
 }
 -> fluent_bit::output{"test":
     config => {
         "elasticsearch" => {
             "host" => "fake.eu-west-1.es.amazonaws.com",
             "port" => #{$mock_es_port},
             "tls" => false,
             "logstashPrefix" => "#{$magic_prefix}",
             "amazonESProxy" => {
               "port" => 9201
             },
           },
         "types" => ["all"],
     },
 }
 EOS
     pp2
   end

   let(:pp3) do
     pp3 = <<-EOS
 host { 'fake.eu-west-1.es.amazonaws.com':
   ensure => present,
   ip => $facts['networking']['dhcp'],
 }
 -> fluent_bit::output{"test":
     config => {
         "types" => ["all"],
     },
 }
 EOS
     pp3
   end

  before(:all) do
    hosts.each do |host|
      # clear firewall
      on host, 'iptables -F INPUT'

      # make sure golang uses proper hosts file
      on host, 'echo "hosts: files dns" > /etc/nsswitch.conf'
    end

    # listen to local ES port
    $mock_es = WEBrick::HTTPServer.new(Port: $mock_es_port)
    $mock_es.mount '/', ESMock
    trap('INT') { $mock_es.shutdown }
    $mock_es_thread = Thread.start { $mock_es.start }
  end

  after(:all) do
    $mock_es.shutdown
    $mock_es_thread.join
  end

  it 'should setup fluent bit without errors based on the example' do
    hosts.each do |host|
      apply_manifest_on(host, pp, catch_failures: true)
      expect(
        apply_manifest_on(host, pp, catch_failures: true).exit_code
      ).to be_zero
    end
  end

  it 'should receive magic log line' do
    $magic_logline_appeared = false
    $magic_prefix_appeared = false

    hosts.each do |host|
      on host, "logger \"#{$magic_logline}\""
    end
    sleep(5)

    expect($magic_logline_appeared).to equal(true)
    expect($magic_prefix_appeared).to equal(false)
  end

  it 'should update fluent bit without errors based on the example' do
    hosts.each do |host|
      apply_manifest_on(host, pp2, catch_failures: true)
      expect(
        apply_manifest_on(host, pp2, catch_failures: true).exit_code
      ).to be_zero
    end
  end

  it 'should receive magic log line and prefix' do
    $magic_logline_appeared = false
    $magic_prefix_appeared = false

    hosts.each do |host|
      on host, "logger \"#{$magic_logline}\""
    end
    sleep(5)

    expect($magic_logline_appeared).to equal(true)
    expect($magic_prefix_appeared).to equal(true)
  end

  it 'should update fluent bit to no ES without errors based on the example' do
    hosts.each do |host|
      apply_manifest_on(host, pp3, catch_failures: true)
      expect(
        apply_manifest_on(host, pp3, catch_failures: true).exit_code
      ).to be_zero
    end
  end

  it 'should not receive magic log line and prefix' do
    $magic_logline_appeared = false
    $magic_prefix_appeared = false

    hosts.each do |host|
      on host, "logger \"#{$magic_logline}\""
    end
    sleep(5)

    expect($magic_logline_appeared).to equal(false)
    expect($magic_prefix_appeared).to equal(false)
  end
end
