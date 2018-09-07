require 'spec_helper'

describe 'role: vault' do
  let(:facts) do
    {
      :tarmak_role          => 'vault',
      :tarmak_type_instance => 'vault',
    }
  end

  it 'sets up airworthy' do
    is_expected.to contain_class('airworthy')
  end

  it 'sets up consul' do
    is_expected.to contain_class('consul')
  end

  it 'sets up vault_server' do
    is_expected.to contain_class('vault_server')
  end

end
