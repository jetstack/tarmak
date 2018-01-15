require 'spec_helper'

describe 'prometheus::server' do
  let(:pre_condition) {[
    'class kubernetes::apiserver{}',
    'require kubernetes::apiserver',
    "concat { '/tmp/spec': owner => 'root', group => 'root', mode  => '0644'}",
    "concat::fragment{'tmpfile': target  => '/tmp/spec', content => 'test contents', order   => '01'}"
  ]}

  let :manifests_file do
    '/etc/kubernetes/apply/prometheus-server.yaml'
  end

  let :manifests_config_file do
    '/etc/kubernetes/apply/prometheus-config.yaml'
  end

  let :rules_config_file do
    '/etc/kubernetes/apply/prometheus-rules.yaml'
  end

  let :config_dir do
    Dir.mktmpdir("prometheus-server")
  end

  let :kubernetes_token_file do
    path = File.join(config_dir, 'kubernetes-token-file')
    File.open(path, 'w') { |file| file.write("i am a token file") }
    path
  end

  let :kubernetes_ca_file do
    path = File.join(config_dir, 'kubernetes-ca-file')
    File.open(path, 'w') { |file| file.write("i am a ca file") }
    path
  end

  let(:params) { {
    'kubernetes_token_file' => kubernetes_token_file,
    'kubernetes_ca_file' => kubernetes_ca_file,
  } }

  let(:manifests) do
    catalogue.resource('Kubernetes::Apply', 'prometheus-server').send(:parameters)[:manifests]
  end

  let(:config_manifest) do
    concat = catalogue.resource('concat_file', manifests_config_file)
    concat.to_ral.should_content
  end

  let(:rules_manifest) do
    concat = catalogue.resource('concat_file', rules_config_file)
    concat.to_ral.should_content
  end


  context 'with default values for all parameters' do
    it 'is valid yaml' do
      manifests.each do |manifest|
        YAML.parse manifest
      end
    end

    it 'should have an emptyDir volume' do
      should contain_file(manifests_file).with_content(/emptyDir: {}/)
    end

    context 'prometheus config' do
      it 'is valid yaml' do
        YAML.parse config_manifest
      end

      it 'is a valid prometheus config' do
        skip 'no promtool found in PATH' unless promtool_available?
        config = YAML.load config_manifest
        file = Tempfile.new('prometheus-config')
        file.write(config['data']['prometheus.yaml'])
        file.close
        Open3.popen3('promtool', 'check', 'config', file.path) do |i,o,e,t|
          i.close
          expect(t.value.exitstatus).to eq(0), "validation of prometheus config had errors: #{o.read} #{e.read}"
        end
      end
    end

    context 'prometheus rules' do
      it 'are valid yaml' do
        YAML.parse rules_manifest
      end

      it 'is a valid prometheus config' do
        skip 'no promtool found in PATH' unless promtool_available?
        config = YAML.load rules_manifest
        file = Tempfile.new('prometheus-rules')
        config['data'].each do |key, value| 
          file.write("---\n")
          file.write(value)
        end
        file.close
        Open3.popen3('promtool', 'check', 'rules', file.path) do |i,o,e,t|
          i.close
          expect(t.value.exitstatus).to eq(0), "validation of prometheus rules had errors: #{o.read} #{e.read}"
        end
      end
    end
  end

  context 'with persistent volume' do
    let(:params) { {
      'persistent_volume' => true,
      'persistent_volume_size' => 13,
    } }

    it 'is valid yaml' do
      manifests.each do |manifest|
        YAML.parse manifest
      end
    end

    it 'should have an persistent volume claim' do
      should_not contain_file(manifests_file).with_content(/emptyDir: {}/)
      should contain_file(manifests_file).with_content(/persistentVolumeClaim:/)
      should contain_file(manifests_file).with_content(/kind: PersistentVolumeClaim/)
      should contain_file(manifests_file).with_content(/storage: 13Gi/)
    end
  end
end
