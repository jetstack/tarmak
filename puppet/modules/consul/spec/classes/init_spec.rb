require 'spec_helper'

describe 'consul' do
    let :app_name do
        'consul'
    end

    let :lib_dir do
        '/var/lib/consul'
    end

    context 'with default values for all parameters' do
        it { should contain_class('consul') }

        it 'should create consul user' do
            should contain_file(lib_dir).with(
                :mode   => '0750',
            )
            should contain_user(app_name).with(
                :home => lib_dir,
            )
        end
    end
end
