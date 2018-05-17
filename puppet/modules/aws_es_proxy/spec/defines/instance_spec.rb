require 'spec_helper'

describe 'aws_es_proxy::instance', :type => :define do
  let(:aws_es_proxy_service_unit) {
    contain_file('/etc/systemd/system/aws-es-proxy-test.service')
  }

  context 'elasticsearch cluster' do
    let(:title) { 'test' }
    let(:params) {
      {
        'dest_address' => 'elastic.example.com',
      }
    }

    it do
      should contain_class('aws_es_proxy')
    end

    it 'should configure aws-es-proxy service unit right' do
      should aws_es_proxy_service_unit.with_content(/#{Regexp.escape('ExecStart=/opt/aws-es-proxy-0.8/aws-es-proxy')}/)
      should aws_es_proxy_service_unit.with_content(/#{Regexp.escape('-endpoint https://elastic.example.com')}/)
      should aws_es_proxy_service_unit.with_content(/#{Regexp.escape('-listen localhost:9200')}/)
    end
  end
end
