require 'spec_helper'

describe 'fragmenthash2yaml' do
  it { is_expected.not_to eq(nil) }
  it { is_expected.to run.with_params().and_raise_error(Puppet::ParseError, /requires one and only one argument/) }
  it { is_expected.to run.with_params({}, {}, {}).and_raise_error(Puppet::ParseError, /requires one and only one argument/) }
  it { is_expected.to run.with_params('some string').and_raise_error(Puppet::ParseError, /requires a hash as argument/) }

  example_input = {
    'domain' => 'example.com',
    'mysql'  => {
      'hosts' => ['192.0.2.2', '192.0.2.4'],
      'user'  => 'root',
      'pass'  => 'setec-astronomy',
    },
    'awesome'  => true,
  }

  context 'default setting' do
    if Puppet.version.to_f < 4.0
      output=<<-EOS
  domain: example.com
  mysql: 
    hosts: 
      - "192.0.2.2"
      - "192.0.2.4"
    user: root
    pass: setec-astronomy
  awesome: true
EOS
    else
      output=<<-EOS
domain: example.com
mysql:
  hosts:
  - 192.0.2.2
  - 192.0.2.4
  user: root
  pass: setec-astronomy
awesome: true
EOS
    end

    it { is_expected.to run.with_params(example_input).and_return(output) }
  end
end
