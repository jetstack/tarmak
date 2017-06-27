require 'spec_helper'

describe 'concat::fragment', :type => :define do

  shared_examples 'fragment' do |title, params|
    params = {} if params.nil?

    p = {
      :content => nil,
      :source  => nil,
      :order   => 10,
    }.merge(params)

    id               = 'root'
    gid              = 'root'

    let(:title) { title }
    let(:params) { params }
    let(:pre_condition) do
      "concat{ '#{p[:target]}': }"
    end

    it do
      should contain_concat(p[:target])
      should contain_concat_file(p[:target])
      should contain_concat_fragment(title)
    end
  end

  context 'title' do
    ['0', '1', 'a', 'z'].each do |title|
      it_behaves_like 'fragment', title, {
        :target  => '/etc/motd',
        :content => "content for #{title}"
      }
    end
  end # title

  context 'target =>' do
    ['./etc/motd', 'etc/motd', 'motd_header'].each do |target|
      context target do
        it_behaves_like 'fragment', target, {
          :target  => '/etc/motd',
          :content => "content for #{target}"
        }
      end
    end

    context 'false' do
      let(:title) { 'motd_header' }
      let(:params) {{ :target => false }}

      it 'should fail' do
        expect { catalogue }.to raise_error(Puppet::Error, /is not a string/)
      end
    end
  end # target =>

  context 'content =>' do
    ['', 'ashp is our hero'].each do |content|
      context content do
        it_behaves_like 'fragment', 'motd_header', {
          :content => content,
          :target  => '/etc/motd',
        }
      end
    end

    context 'false' do
      let(:title) { 'motd_header' }
      let(:params) {{ :content => false, :target => '/etc/motd' }}

      it 'should fail' do
        expect { catalogue }.to raise_error(Puppet::Error, /is not a string/)
      end
    end
  end # content =>

  context 'source =>' do
    ['', '/foo/bar', ['/foo/bar', '/foo/baz']].each do |source|
      context source do
        it_behaves_like 'fragment', 'motd_header', {
          :source => source,
          :target => '/etc/motd',
        }
      end
    end

    context 'false' do
      let(:title) { 'motd_header' }
      let(:params) {{ :source => false, :target => '/etc/motd' }}

      it 'should fail' do
        expect { catalogue }.to raise_error(Puppet::Error, /is not a String or an Array/)
      end
    end
  end # source =>

  context 'order =>' do
    ['', '42', 'a', 'z'].each do |order|
      context '\'\'' do
        it_behaves_like 'fragment', 'motd_header', {
          :order  => order,
          :target => '/etc/motd',
        }
      end
    end

    context 'false' do
      let(:title) { 'motd_header' }
      let(:params) {{ :order => false, :target => '/etc/motd' }}

      it 'should fail' do
        expect { catalogue }.to raise_error(Puppet::Error, /is not a String or an Integer/)
      end
    end

    context '123:456' do
      let(:title) { 'motd_header' }
      let(:params) {{ :order => '123:456', :target => '/etc/motd' }}

      it 'should fail' do
        expect { catalogue }.to raise_error(Puppet::Error, /cannot contain/)
      end
    end
    context '123/456' do
      let(:title) { 'motd_header' }
      let(:params) {{ :order => '123/456', :target => '/etc/motd' }}

      it 'should fail' do
        expect { catalogue }.to raise_error(Puppet::Error, /cannot contain/)
      end
    end
    context '123\n456' do
      let(:title) { 'motd_header' }
      let(:params) {{ :order => "123\n456", :target => '/etc/motd' }}

      it 'should fail' do
        expect { catalogue }.to raise_error(Puppet::Error, /cannot contain/)
      end
    end
  end # order =>

  context 'more than one content source' do
    context 'source and content' do
      let(:title) { 'motd_header' }
      let(:params) do
        {
          :target => '/etc/motd',
          :source => '/foo',
          :content => 'bar',
        }
      end

      it 'should fail' do
        expect { catalogue }.to raise_error(Puppet::Error, /Can\'t use \'source\' and \'content\' at the same time/m)
      end
    end
  end # more than one content source
end
