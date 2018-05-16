require 'spec_helper'

describe 'fluent_bit::output', :type => :define do

  let(:config) {
    contain_file('/etc/td-agent-bit/td-agent-bit.conf')
  }

  let(:output) {
    contain_file('/etc/td-agent-bit/td-agent-bit-output-test.conf')
  }

  let(:aws_es_proxy_service_unit) {
    contain_file('/etc/systemd/system/aws-es-proxy-test.service')
  }

  context 'elasticsearch cluster' do
    let(:title) { 'test' }
    let(:params) {
      {
        :config => {"elasticsearch" => {
            "host" => "elastic.example.com",
            "port" => 443,
            "tls" => true,
            "tlsVerify" => true,
        },
          "types" => ["all"],
        },
      }
    }

    it do
      should contain_class('fluent_bit')
    end

    it do
      should_not contain_class('aws_es_proxy')
    end

    it 'should configure output right' do
      should output.with_content(/#{Regexp.escape('tls On')}/)
      should output.with_content(/#{Regexp.escape('tls.verify On')}/)
      should output.with_content(/#{Regexp.escape('Host elastic.example.com')}/)
    end

  end

  context 'elasticsearch cluster with aws-es-proxy' do
    let(:title) { 'test' }
    let(:params) {
      {
        :config => {"elasticsearch" => {
            "host" => "elastic.example.com",
            "port" => 443,
            "tls" => true,
            "tlsVerify" => true,
            "amazonESProxy" => {
              "port" => 9201
            },
          },
          "types" => ["all"],
        },
      }
    }

    it do
      should contain_class('fluent_bit')
    end

    it do
      should contain_class('aws_es_proxy')
    end

    it 'should configure output right' do
      should output.without_content(/#{Regexp.escape('tls On')}/)
      should output.without_content(/#{Regexp.escape('tls.verify On')}/)
      should output.with_content(/#{Regexp.escape('Host 127.0.0.1')}/)
      should output.with_content(/#{Regexp.escape('Port 9201')}/)
    end

    it 'should configure aws-es-proxy service unit right' do
      should aws_es_proxy_service_unit.with_content(/#{Regexp.escape('ExecStart=/opt/aws-es-proxy-0.8/aws-es-proxy')}/)
      should aws_es_proxy_service_unit.with_content(/#{Regexp.escape('-endpoint https://elastic.example.com')}/)
      should aws_es_proxy_service_unit.with_content(/#{Regexp.escape('-listen localhost:9201')}/)
    end

  end

end
