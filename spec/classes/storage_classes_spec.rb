require 'spec_helper'

describe 'kubernetes::storage_classes' do
  let(:pre_condition) do
    [
      "class{'kubernetes':
         version => '#{version}',
         cloud_provider => '#{cloud_provider}',
      }",
      'include kubernetes::apiserver'
    ]
  end

  let :cloud_provider do
    ''
  end

  let :version do
    '1.4.8'
  end

  let :service_file do
    '/etc/systemd/system/kubectl-apply-storage-classes.service'
  end

  let :manifests_file do
    '/etc/kubernetes/apply/storage-classes.yaml'
  end

  context 'no cloud_provider' do
    it 'should no contain storage classes' do
      should_not contain_file(manifests_file).with_content(/aws-ebs/)
      should_not contain_file(service_file)
    end
  end

  context 'cloud_provider = aws' do
    let :cloud_provider do
      'aws'
    end

    context 'kubernetes < 1.4' do
      let :version do
        '1.3.9'
      end
      it 'should no contain storage classes' do
        should_not contain_file(manifests_file).with_content(/aws-ebs/)
        should_not contain_file(service_file)
      end
    end

    context 'kubernetes >= 1.4' do
      it 'should contain aws storage classes' do
        should contain_file(manifests_file).with_content(/aws-ebs/)
        should contain_file(service_file)
      end
    end
  end
end
