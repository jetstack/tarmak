require 'logger'
require 'open3'
require 'json'
require 'yaml'
require 'base64'
require 'rhcl'

logger = Logger.new(STDERR)
logger.level = Logger::DEBUG

namespace :aws do
  task :prepare do
    require 'aws-sdk'
    require 'inifile'
    @aws_region = ENV['AWS_DEFAULT_REGION'] || 'eu-west-1'
    @aws_profile = ENV['AWS_PROFILE'] || 'ss_non_prod'
    @aws_config_file = IniFile.load(ENV['HOME'] + '/.aws/config')
    begin
      @aws_config = @aws_config_file["profile #{@aws_profile}"]
    rescue
      @aws_config = {}
    end
    @aws_credentials = Aws::SharedCredentials.new(profile_name: @aws_profile)
  end

  desc 'login using jetstack vault'
  task :login_jetstack do
    cmd = ['vault', 'read', '-format', 'json', 'jetstack/aws/jetstack-dev/sts/admin']
    Open3.popen3(*cmd) do | stdin, stdout, stderr, wait_thr|
      stdin.close
      fail "Getting credentails from vault failed: #{stderr.read}" if wait_thr.value != 0
      credentials = JSON.parse(stdout.read)
      puts "export AWS_ACCESS_KEY_ID=#{credentials['data']['access_key']}"
      puts "export AWS_SECRET_ACCESS_KEY=#{credentials['data']['secret_key']}"
      puts "export AWS_SESSION_TOKEN=#{credentials['data']['security_token']}"
    end
  end
end

namespace :terraform do

  task :prepare_env => :'aws:prepare' do
    @terraform_plan= ENV['TERRAFORM_PLAN']
    @terraform_environments = ['nonprod']
    @terraform_environment = ENV['TERRAFORM_ENVIRONMENT'] || 'nonprod'
    unless @terraform_environments.include?(@terraform_environment)
      fail "Please provide a TERRAFORM_ENVIRONMENT out of #{@terraform_environments}"
    end
    tfvars = Rhcl.parse(File.open("tfvars/network_#{@terraform_environment}_hub.tfvars").read)
    @terraform_state_bucket = "#{tfvars['bucket_prefix']}#{@terraform_environment}-#{@aws_region}-terraform-state"
    @terraform_names = /^[a-z0-9]{3,16}$/
    if not @terraform_names.match(ENV['TERRAFORM_NAME'])
      fail "Please provide a TERRAFORM_NAME variable with that matches #{@terraform_names}"
    end
    @terraform_name = ENV['TERRAFORM_NAME']
  end

  task :prepare => :prepare_env do
    @terraform_stacks = ['network', 'vault', 'tools', 'kubernetes']
    unless @terraform_stacks.include?(ENV['TERRAFORM_STACK'])
      fail "Please provide a TERRAFORM_STACK out of #{@terraform_stacks}"
    end
    @terraform_stack = ENV['TERRAFORM_STACK']

    sh 'mkdir -p tfstate'
    terraform_file_base = "#{@terraform_stack}_#{@terraform_environment}_#{@terraform_name}"
    @terraform_state_file = "#{terraform_file_base}.tfstate"
    @terraform_vars_file = "#{terraform_file_base}.tfvars"

    @terraform_args = [
      "-var-file=../tfvars/#{@terraform_vars_file}"
    ] + ['name', 'environment', 'stack','state_bucket'].map do |name|
      "-var=#{name}=#{instance_variable_get(("@terraform_" + name))}"
    end

    # configure remote state
    Dir.chdir(@terraform_stack) do
      sh "rm -rf .terraform"
      sh "terraform remote config -backend=s3 '-backend-config=bucket=#{@terraform_state_bucket}' '-backend-config=key=#{@terraform_state_file}' '-backend-config=region=#{@aws_region}'"
    end
  end

  task :hub_outputs => :prepare_env do
    @s3 = Aws::S3::Resource.new(region: @aws_region)
    bucket = @s3.bucket(@terraform_state_bucket)
    state = JSON.parse(bucket.object("network_#{@terraform_environment}_hub.tfstate").get.body.read)
    state['modules'].each do |mod|
      next if mod['path'] != ["root"]
      @terraform_hub_outputs = mod['outputs']
    end
    fail "No hub outputs found" if @terraform_hub_outputs.nil?
  end

  task :plan => :prepare do
    Dir.chdir(@terraform_stack) do
      args = @terraform_args
      args << '-var-file=/work/tokens.tfvar' if @terraform_stack == 'kubernetes' and File.exists?('/work/tokens.tfvar')
      # generate plan and return a 2 exitcode if there's something to change
      if not @terraform_plan.nil?
        args << "-out=#{@terraform_plan}"
        args << '-detailed-exitcode'
        args << '-destroy' if ENV['TERRAFORM_DESTROY'] == 'true'
        sh 'terraform', 'plan', *args do |ok, res|
          fail "terraform plan failed" if res.exitstatus != 0 and res.exitstatus != 2
          exit res.exitstatus
        end
      else
        sh 'terraform', 'plan', *args
      end
    end
  end

  task :apply => :prepare do
    Dir.chdir(@terraform_stack) do
      if @terraform_plan.nil?
        args = @terraform_args
      else
        args = [@terraform_plan]
      end
      sh 'terraform', 'apply', *args
    end
  end

  task :destroy => :prepare do
    Dir.chdir(@terraform_stack) do
      args = ['terraform', 'destroy'] + @terraform_args
      args << '-force' if ENV['TERRAFORM_DESTROY']
      sh(*args)
    end
  end

  task :fmt do
    sh 'find . -name "*.tf" | xargs -n1 dirname | sort -u | xargs -n 1 terraform fmt -write=false -diff=true'
  end

  task :validate do
    sh 'find . -name "*.tf" | xargs -n1 dirname | sort -u | xargs -n 1 terraform validate'
  end
end

namespace :packer do
  task :build do
    Dir.chdir('packer') do
    sh 'packer', 'build', "#{ENV['PACKER_NAME']}.json"
    end
  end
end

namespace :vault do
  task :prepare => :'terraform:hub_outputs' do
    vault_instances = ENV['VAULT_INSTANCES'] || 5
    @vault_instances = vault_instances.to_i
    @vault_zone = @terraform_hub_outputs['private_zones']['value'].first
    @vault_path = "vault-#{@terraform_environment}"
    logger.info "vault CA zone=#{@vault_zone} instances=#{@vault_instances}"

    # generate node names
    @vault_cn = "vault.#{@vault_zone}"
    @vault_nodes = [@vault_cn]
    (1..@vault_instances).to_a.each do |i|
      @vault_nodes << "vault-#{i}.#{@vault_zone}"
    end
    @vault_nodes << 'localhost'

    secrets_bucket = @terraform_hub_outputs['secrets_bucket']['value']
    @secrets_bucket = @s3.bucket(secrets_bucket)
    @secrets_kms_arn = @terraform_hub_outputs['secrets_kms_arn']['value']
    logger.info "secrets bucket=#{secrets_bucket} kms_arn=#{@secrets_kms_arn}"
  end

  desc 'Ensure Cert certificate exists'
  task :secrets_ca => :prepare do
    spec = {
      'CN' => "Vault CA #{@terraform_environment}",
      'key' => { 'algo' => 'rsa', 'size' => 2048 },
      'ca' => { 'expiry' => '262800h' }, # expire after 3 years
    }
    cert_path = "#{@vault_path}/ca.pem"
    key_path = "#{@vault_path}/ca-key.pem"

    begin
      ca = {}
      [:cert, :key].each do |type|
        obj = @secrets_bucket.object(instance_eval("#{type.to_s}_path")).get
        ca[type.to_s] = obj.body.read
      end
      @ca = ca
    rescue Aws::S3::Errors::NoSuchKey
      logger.info "Generating a new CA certificate"
      Open3.popen3('cfssl', 'gencert', '-initca', '-') do |stdin, stdout, stderr, wait_thr|
        stdin.write(JSON.generate(spec))
        stdin.close
        fail "Generating CA failed: #{stderr.read}" if wait_thr.value != 0
        @ca = JSON.parse(stdout.read)
        @secrets_bucket.put_object(
          key: cert_path,
          body: @ca['cert'],
          server_side_encryption: 'aws:kms',
          ssekms_key_id: @secrets_kms_arn,
          content_type: 'text/plain',
        )
        @secrets_bucket.put_object(
          key: key_path,
          body: @ca['key'],
          server_side_encryption: 'aws:kms',
          ssekms_key_id: @secrets_kms_arn,
          content_type: 'text/plain',
        )
      end
    end
  end

  desc 'Ensure CA certificate exists'
  task :secrets_cert => :secrets_ca do
    csr = {
      'CN' => @vault_cn,
      'hosts' => @vault_nodes,
      'key' => { 'algo' => 'rsa', 'size' => 2048 },
    }
    ca_config = {
      'signing' => {
        'default' =>  { 'expiry' => '43800h'},
        'profiles' => {
          'server' => {
            'expiry' => '43800h',
            'usages' => ['signing', 'key encipherment', 'server auth'],
          }
        }
      }
    }

    cert_path = "#{@vault_path}/cert.pem"
    key_path = "#{@vault_path}/cert-key.pem"

    begin
      [:cert, :key].each do |type|
        @secrets_bucket.object(instance_eval("#{type.to_s}_path")).get
      end
    rescue Aws::S3::Errors::NoSuchKey
      logger.info "Generating a new certificate"

      temp_files = [
        JSON.generate(ca_config),
        @ca['key'],
        @ca['cert'],
      ].map do |contents|
        file = Tempfile.new
        file.write(contents)
        file.close
        file
      end

      cmd = ['cfssl', 'gencert', "-ca=#{temp_files[2].path}", "-ca-key=#{temp_files[1].path}", "-config=#{temp_files[0].path}", '-profile=server', "-hostname=#{@vault_nodes.join(',')}", '-']

      Open3.popen3(*cmd) do | stdin, stdout, stderr, wait_thr|
        stdin.write(JSON.generate(csr))
        stdin.close
        fail "Generating cert failed: #{stderr.read}" if wait_thr.value != 0
        cert = JSON.parse(stdout.read)
        @secrets_bucket.put_object(
          key: cert_path,
          body: cert['cert'],
          server_side_encryption: 'aws:kms',
          ssekms_key_id: @secrets_kms_arn,
          content_type: 'text/plain',
        )
        @secrets_bucket.put_object(
          key: key_path,
          body: cert['key'],
          server_side_encryption: 'aws:kms',
          ssekms_key_id: @secrets_kms_arn,
          content_type: 'text/plain',
        )
      end

      # cleanup files
      temp_files.each(&:unlink)
    end
  end

  task :secrets => [:secrets_ca, :secrets_cert]

  desc 'Initialize vault if needed'
  task :initialize => :prepare do
    url = "https://#{@vault_cn}:8200"
    logger.info "vault url = #{url}"
    uri = URI(url)

    # retry initialize for 10 times
    retries = 100
    begin
      resp = Net::HTTP.start(
        uri.host, uri.port,
        :use_ssl => uri.scheme == 'https',
        :verify_mode => OpenSSL::SSL::VERIFY_NONE,
      ) do |http|
        resp = JSON.parse(http.request(Net::HTTP::Get.new('/v1/sys/init')).body)

        if resp['initialized']
          logger.info 'vault is already initialized'
        else
          logger.info 'initialize vault'

          req = Net::HTTP::Put.new('/v1/sys/init', { 'Content-Type' => 'application/json'})
          req.body = JSON.generate({:secret_shares => 1, :secret_threshold => 1})

          resp = JSON.parse(http.request(req).body)

          logger.debug 'store root token in S3'
          @secrets_bucket.put_object(
            key: "#{@vault_path}/root-token",
            body: resp['root_token'],
            server_side_encryption: 'aws:kms',
            ssekms_key_id: @secrets_kms_arn,
          )

          logger.debug 'store unseal key in AWS parameter store'
          ssm = Aws::SSM::Client.new(region: 'eu-west-1')
          ssm.put_parameter({
            name: "vault-#{@terraform_environment}-unseal-key",
            value: resp['keys_base64'].first,
            type: 'SecureString',
            key_id: @secrets_kms_arn,
            overwrite: true,
          })
        end
      end
    rescue Errno::ECONNREFUSED => e
      retries -= 1
      if retries > 0
        logger.warn 'Connection to vault failed, retrying in 5 seconds'
        sleep 5
        retry
      else
        raise e
      end
    end
  end

  task :prepare_login => :prepare do
    ca_s3_path = "#{@vault_path}/ca.pem"
    root_token_s3_path = "#{@vault_path}/root-token"
    ca_file = Tempfile.new
    ca_file.write(@secrets_bucket.object(ca_s3_path).get.body.read)
    ca_file.close
    root_token = @secrets_bucket.object(root_token_s3_path).get.body.read
    puts root_token
    ENV['VAULT_ADDR'] = "https://#{@vault_cn}:8200"
    ENV['VAULT_TOKEN'] = root_token
    ENV['VAULT_CACERT'] = ca_file.path
    @terraform_name = ENV['TERRAFORM_NAME']
    @cluster_name = "#{@terraform_environment}-#{@terraform_name}"
  end

  desc 'Setup a k8s cluster in vault'
  task :setup_k8s => :prepare_login do
    ENV['CLUSTER_ID'] = @cluster_name
    sh "vault/scripts/setup_vault.sh"
  end

  desc 'Generate kubeconfig for cluster'
  task :kubeconfig => :prepare_login do
    kubeconfig = {
      'current-context' => @cluster_name,
      'apiVersion' => 'v1',
      'clusters' => [{
        'cluster' => {
          'apiVersion' => 'v1',
          'server' => 'https://localhost:6443',
        },
        'name' => @cluster_name,
      }],
      'contexts' => [{
        'context' => {
          'cluster' => @cluster_name,
          'namespace' => 'kube-system',
          'user' => @cluster_name,
        },
        'name' => @cluster_name,
      }],
      'kind' => 'Config',
      'preferences' => {
        'colors' => true,
      },
      'users' => [{
        'name' => @cluster_name,
        'user' => {},
      }],
    }
    api_host = "api.#{@cluster_name}.#{@terraform_hub_outputs['private_zones']['value'].first}:6443"
    tunnel_host = "localhost:6443"
    cmd = ['vault', 'write', '-format', 'json', "#{ENV['CLUSTER_ID']}/nonprod-devcluster/pki/k8s/issue/admin", "common_name=admin"]
    Open3.popen3(*cmd) do | stdin, stdout, stderr, wait_thr|
      stdin.close
      fail "Getting credentails from vault failed: #{stderr.read}" if wait_thr.value != 0
      creds = JSON.parse(stdout.read)['data']
      kubeconfig['users'][0]['user']['client-key-data'] = Base64.encode64(creds['private_key'])
      kubeconfig['users'][0]['user']['client-certificate-data'] = Base64.encode64(creds['certificate'])
      kubeconfig['clusters'][0]['cluster']['certificate-authority-data'] = Base64.encode64(creds['issuing_ca'])
      dest_file = 'kubeconfig-tunnel'
      File.open(dest_file, 'w') do |f|
        f.write "# SSH tunnel to API via Bastion:\n"
        f.write "# ssh -N -L6443:#{api_host} centos@bastion.#{@terraform_hub_outputs['public_zones']['value'].first}\n"
        f.write "#\n\n"
        f.write kubeconfig.to_yaml
      end
      logger.info "Wrote #{dest_file}"
      dest_file = 'kubeconfig-private'
      File.open(dest_file, 'w') do |f|
        kubeconfig['clusters'][0]['cluster']['server'] = "https://#{api_host}"
        f.write kubeconfig.to_yaml
      end
      logger.info "Wrote #{dest_file}"
    end
  end
end

namespace :puppet do
  task :prepare => :'terraform:hub_outputs' do
    zone = @terraform_hub_outputs['private_zones']['value'].first
    @puppet_master = "puppet.#{zone}"
  end

  desc 'Deploy puppet.tar.gz to the puppet master'
  task :deploy_env => :prepare do
    sh "cat puppet.tar.gz | ssh -o StrictHostKeyChecking=no puppet-deploy@#{@puppet_master} #{@terraform_environment}_#{@terraform_name}"
  end
end
