require 'logger'
require 'open3'
require 'json'
require 'yaml'
require 'base64'
require 'rhcl'
require 'excon'
require 'securerandom'

logger = Logger.new(STDERR)
logger.level = Logger::DEBUG

def instance_name(i)
  i.tags.each do |tag|
    return tag.value if tag.key == 'Name'
  end
  return 'unknown'
end

class Jenkins
    JOB_XML = <<-EOS
<flow-definition plugin="workflow-job@2.10">
  <actions/>
  <description/>
  <keepDependencies>false</keepDependencies>
  <properties/>
  <definition class="org.jenkinsci.plugins.workflow.cps.CpsScmFlowDefinition" plugin="workflow-cps@2.29">
    <scm class="hudson.plugins.git.GitSCM" plugin="git@3.1.0">
      <configVersion>2</configVersion>
      <userRemoteConfigs>
        <hudson.plugins.git.UserRemoteConfig>
          <url>%<git_url>s</url>
          <credentialsId>%<git_secret>s</credentialsId>
        </hudson.plugins.git.UserRemoteConfig>
      </userRemoteConfigs>
      <branches>
        <hudson.plugins.git.BranchSpec>
          <name>%<git_branch>s</name>
        </hudson.plugins.git.BranchSpec>
      </branches>
      <doGenerateSubmoduleConfigurations>false</doGenerateSubmoduleConfigurations>
      <submoduleCfg class="list"/>
      <extensions/>
    </scm>
    <scriptPath>%<jenkinsfile>s</scriptPath>
    <lightweight>true</lightweight>
  </definition>
  <triggers/>
</flow-definition>
EOS
    FOLDER_XML = <<-EOS
<?xml version="1.0" encoding="UTF-8"?>
<com.cloudbees.hudson.plugins.folder.Folder plugin="cloudbees-folder@5.18">
   <actions />
   <description />
   <properties />
   <folderViews class="com.cloudbees.hudson.plugins.folder.views.DefaultFolderViewHolder">
      <views>
         <hudson.model.AllView>
            <owner class="com.cloudbees.hudson.plugins.folder.Folder" reference="../../../.." />
            <name>All</name>
            <filterExecutors>false</filterExecutors>
            <filterQueue>false</filterQueue>
            <properties class="hudson.model.View$PropertyList" />
         </hudson.model.AllView>
      </views>
      <tabBar class="hudson.views.DefaultViewsTabBar" />
   </folderViews>
   <healthMetrics>
      <com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric>
         <nonRecursive>false</nonRecursive>
      </com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric>
   </healthMetrics>
   <icon class="com.cloudbees.hudson.plugins.folder.icons.StockFolderIcon" />
</com.cloudbees.hudson.plugins.folder.Folder>
EOS

  def initialize(*args)
    @client = Excon.new(*args)
    @logger = Logger.new(STDERR)
  end

  def post(opts)
    crub = @client.get(
      :expects => [200, 201],
      :path => '/crumbIssuer/api/json',
    )
    crumb = JSON.parse(crub.body)
    opts[:headers] = {} if opts[:headers].nil?
    opts[:headers][crumb['crumbRequestField']] = crumb['crumb']

    @client.post(opts)
  end

  def create_split_path(in_path)
    parts = in_path.split('/')

    path = ''
    parts[0,parts.length-1].each do |part|
      path = File.join path, 'job', part
    end
    path = File.join path, '/createItem'

    return parts.last, path
  end

  def create_job(job, opts={})
    opts = {
      git_branch: '*/master',
      jenkinsfile: 'Jenkinsfile',
    }.update(opts)

    name, path = create_split_path(job)
    begin
      resp = post(
        :expects => [200, 201],
        :path => path,
        :query => {'name' => name},
        :body => sprintf(JOB_XML, opts),
        :headers => {
          "Content-Type" => "application/xml",
        },
      )
      @logger.debug "job #{job} created"
      resp
    rescue Excon::Error::BadRequest => e
      if e.response.headers['X-Error'].match(/job already exists/)
        @logger.debug "job #{job} already exists"
        e.response
      else
        raise e
      end
    end
  end

  def create_folder(folder)
    name, path = create_split_path(folder)
    begin
      resp = post(
        :expects => [200, 201],
        :path => path,
        :query => {'name' => name},
        :body => FOLDER_XML,
        :headers => {
          "Content-Type" => "application/xml",
        },
      )
      @logger.debug "folder #{folder} created"
      resp
    rescue Excon::Error::BadRequest => e
      if e.response.headers['X-Error'].match(/job already exists/)
        @logger.debug "folder #{folder} already exists"
        e.response
      else
        raise e
      end
    end
  end

  def create_ssh_credential(name, username, private_key)
    json = {
      credentials: {
        username: username,
        'privateKeySource': {
          value: '0',
          'privateKey': private_key,
          'stapler-class': 'com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey$DirectEntryPrivateKeySource',
        },
        passphrase: '',
        id: name,
        description: "puppernetes secret #{name}",
        'stapler-class': 'com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey',
        '$class': 'com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey',
      },
    }

    resp = post(
      :expects => [301, 302],
      :path => '/credentials/store/system/domain/_/createCredentials',
      :body => URI.encode_www_form(:json =>  JSON.generate(json)),
      :headers => {
        'Content-Type' => 'application/x-www-form-urlencoded',
      },
    )
    @logger.debug "ensured ssh credentials #{name} exists"
    resp
  end

  def create_file_credential(name, private_key)
    body      = ''
    boundary  = SecureRandom.hex(8)

    json = {
      credentials: {
        file: 'file0',
        id: name,
        description: "puppernetes file secret #{name}",
        'stapler-class': 'org.jenkinsci.plugins.plaincredentials.impl.FileCredentialsImpl',
        '$class': 'org.jenkinsci.plugins.plaincredentials.impl.FileCredentialsImpl'
      },
    }
    body << "--#{boundary}" << Excon::CR_NL
    body << 'Content-Disposition: form-data; name="file0"; filename="ssh_id"' << Excon::CR_NL
    body << 'Content-Type: text/plain' << Excon::CR_NL
    body << Excon::CR_NL
    body << private_key
    body << Excon::CR_NL
    body << "--#{boundary}" << Excon::CR_NL
    body << 'Content-Disposition: form-data; name="json"' << Excon::CR_NL
    body << Excon::CR_NL
    body << JSON.generate(json)
    body << Excon::CR_NL
    body << "--#{boundary}--" << Excon::CR_NL

    resp = post(
      :expects => [301, 302],
      :path => '/credentials/store/system/domain/_/createCredentials',
      :body => body,
      :headers => {
        'Content-Type' => %{multipart/form-data; boundary="#{boundary}"},
      },
    )
    @logger.debug "ensured file credentials #{name} exists"
    resp
  end

  def install_plugins(*plugins)
    xml = '<jenkins>'
    plugins.each do |plugin|
      xml << %{<install plugin="#{plugin}" />}
    end
    xml << '</jenkins>'
    post(
      :expects => [301, 302],
      :path => '/pluginManager/installNecessaryPlugins',
      :body => xml,
      :headers => {
        'Content-Type' => 'application/xml',
      },
    )
    @logger.debug "installing plugins #{plugins.join(',')}"
  end

  def safe_restart
    post(
      :expects => [301, 302],
      :path => '/updateCenter/safeRestart',
    )
    @logger.debug "requesting a jenkins restart"
  end
end

namespace :aws do
  task :prepare => [:'terraform:global_tfvars'] do
    require 'aws-sdk'
    require 'inifile'
    @aws_region = @terraform_global_tfvars['region']
    ENV['AWS_DEFAULT_REGION'] = @aws_region
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
  task :login_jetstack => [:'terraform:global_tfvars'] do
    cmd = ['vault', 'read', '-format', 'json', File.join(@terraform_global_tfvars['jetstack_vault_aws_path'],'sts/admin')]
    Open3.popen3(*cmd) do | stdin, stdout, stderr, wait_thr|
      stdin.close
      fail "Getting credentails from vault failed: #{stderr.read}" if wait_thr.value != 0
      credentials = JSON.parse(stdout.read)
      puts "export AWS_ACCESS_KEY_ID=#{credentials['data']['access_key']}"
      puts "export AWS_SECRET_ACCESS_KEY=#{credentials['data']['secret_key']}"
      puts "export AWS_SESSION_TOKEN=#{credentials['data']['security_token']}"
    end
  end

  desc 'ensure EC2 key pair exists'
  task :ensure_key_pair => [:prepare, :'terraform:global_tfvars'] do
    ec2 = Aws::EC2::Client.new(region: @aws_region)
    key_name = @terraform_global_tfvars['key_name']
    begin
      ec2.describe_key_pairs({key_names:[key_name]})
      logger.info "AWS key pair #{key_name} already exists"
    rescue Aws::EC2::Errors::InvalidKeyPairNotFound
      key_pair_path = 'credentials/aws_key_pair'

      logger.info 'generating new AWS key pair'
      sh 'mkdir', '-p', File.dirname(key_pair_path)
      sh 'ssh-keygen', '-t', 'rsa', '-b', '4096', '-N', '', '-f', key_pair_path, '-C', "aws-keypair-#{key_name}"

      ec2.import_key_pair({
        key_name: key_name,
        public_key_material: File.open("#{key_pair_path}.pub").read,
      })
      logger.info "AWS key pair #{key_name} generated and public key uploaded"
      logger.warn "Make sure you save the private key in '#{key_pair_path}'"
    end
  end
end

namespace :terraform do
  desc 'parse terraform global variables'
  task :global_tfvars do
    @terraform_global_tfvars = Rhcl.parse(File.open("tfvars/global.tfvars").read)
  end

  task :prepare_regex do
    @terraform_regex = /^[a-z0-9]{3,16}$/
  end

  task :prepare_environment => :prepare_regex do
    key = 'TERRAFORM_ENVIRONMENT'
    if not ENV[key]
      @terraform_environment = 'nonprod'
    elsif not @terraform_regex.match(ENV[key])
      fail "Please provide a #{key} variable with that matches #{@terraform_regex}"
    else
      @terraform_environment = ENV[key]
    end
  end

  task :prepare_name => :prepare_regex do
    key = 'TERRAFORM_NAME'
    if not @terraform_regex.match(ENV[key])
      fail "Please provide a #{key} variable with that matches #{@terraform_regex}"
    else
      @terraform_name = ENV[key]
    end
  end

  task :prepare_env => [:'aws:prepare', :prepare_name, :prepare_environment] do
    @terraform_plan= ENV['TERRAFORM_PLAN']
    tfvars = Rhcl.parse(File.open("tfvars/#{@terraform_environment}/hub/network.tfvars").read)
    @terraform_state_bucket = "#{tfvars['bucket_prefix']}#{@terraform_environment}-#{@aws_region}-terraform-state"
  end

  task :prepare => :prepare_env do
    @terraform_stacks = ['network', 'vault', 'tools', 'kubernetes']
    unless @terraform_stacks.include?(ENV['TERRAFORM_STACK'])
      fail "Please provide a TERRAFORM_STACK out of #{@terraform_stacks}"
    end
    @terraform_stack = ENV['TERRAFORM_STACK']

    terraform_file_base = "#{@terraform_stack}_#{@terraform_environment}_#{@terraform_name}"
    @terraform_state_file = "#{terraform_file_base}.tfstate"

    @terraform_args = []

    # adds global tfvars
    path = "tfvars/global.tfvars"
    if File.exist? path
      @terraform_args << "-var-file=../#{path}"
    end

    # adds per environment tfvars
    path = "tfvars/#{@terraform_environment}/environment.tfvars"
    if File.exist? path
      @terraform_args << "-var-file=../#{path}"
    end

    # adds per stack
    path = "tfvars/#{@terraform_environment}/#{@terraform_name}/#{@terraform_stack}.tfvars"
    if File.exist? path
      @terraform_args << "-var-file=../#{path}"
    end

    @terraform_args += ['name', 'environment', 'stack','state_bucket'].map do |name|
      "-var=#{name}=#{instance_variable_get(("@terraform_" + name))}"
    end

    # configure remote state
    Dir.chdir(@terraform_stack) do
      terraform_remote_state_file = 'terraform_remote_state.tf'
      if ENV['TERRAFORM_DISABLE_REMOTE_STATE'] != 'true'
        remote_state = [
          'terraform {',
          '  backend "s3" {',
          "    bucket = \"#{@terraform_state_bucket}\"",
          "    key = \"#{@terraform_state_file}\"",
          "    region = \"#{@aws_region}\"",
          "    lock_table = \"#{@terraform_state_bucket}\"",
          ' }',
          '}'
        ].join("\n")
        File.open(terraform_remote_state_file, 'w') do |f|
          f.write remote_state
        end
        sh 'terraform init'
      else
        sh 'rm', '-rf', terraform_remote_state_file
      end
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
          File.open('../.terraform_exitcode', 'w') do |f|
            f.write res.exitstatus.to_s
          end
          if res.exitstatus != 0 and res.exitstatus != 2
            puts "terraform plan failed"
            exit res.exitstatus
          end
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
      # clean up plan, to prevent duplication
      sh 'rm', '-f', @terraform_plan unless @terraform_plan.nil?
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
  task :build => [:'aws:prepare'] do
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
          ssm = Aws::SSM::Client.new(region: @aws_region)
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

  desc 'deploy puppet.tar.gz to the puppet master'
  task :deploy_env => :prepare do
    sh "cat puppet.tar.gz | ssh -o StrictHostKeyChecking=no puppet-deploy@#{@puppet_master} #{@terraform_environment}_#{@terraform_name}"
  end

  desc 'run puppet apply on every node in a cluster'
  task :node_apply => :prepare do
    require 'aws-sdk'
    require 'thread'
    semaphore = Mutex.new
    cluster_id = "#{@terraform_environment}-#{@terraform_name}"
    ec2 = Aws::EC2::Resource.new(region: @aws_region)

    threads = []
    results_failed = []
    results_successful = []

    # Get all instances in cluster
    ec2.instances({filters: [
      {name: 'tag:KubernetesCluster', values: [cluster_id]},
    ]}).each do |i|
      # ignore non running instances
      next if i.state.name != 'running'

      logger.info "connecting to host #{instance_name(i)} (#{i.private_ip_address})"
      threads << Thread.new do
        node = i.private_ip_address
        cmd = [
          'ssh',
          '-o',
          'StrictHostKeyChecking=no',
          '-o',
          'ConnectTimeout=10',
          "puppet-deploy@#{node}",
        ]
        Open3.popen3(*cmd) do |stdin, stdout, stderr, wait_thr|
          stdin.close
          exit_code = wait_thr.value
          output_stdout = stdout.read
          output_stderr = stderr.read
          output = output_stdout
          if stderr != ''
            output += "\nSTDERR: #{output_stderr}"
          end
          semaphore.synchronize do
            if exit_code == 0 or exit_code == 2
              results_successful << [i, output]
            else
              results_failed << [i, output]
            end
          end
        end
      end
    end
    threads.each do |thr|
      thr.join
    end
    return_code = 0

    results_successful.each do |i, output|
      puts "execution on host #{instance_name(i)} (#{i.private_ip_address}) succeeded:"
      puts output
    end
    results_failed.each do |i, output|
      puts "execution on host #{instance_name(i)} (#{i.private_ip_address}) failed:"
      puts output
      return_code = 1
    end
    exit return_code
  end
end

namespace :jenkins do
  task :list_folders => [:list_jobs] do
    # create a list of unique parent folders
    folders = @jenkins_jobs.keys.map do |job, values|
      ret = []
      current = ''
      File.dirname(job.to_s).split('/').each do |part|
        if current.length == 0 
          current = part
        else
          current = File.join(current, part)
        end
        ret << current
      end
      ret
    end.flatten.uniq

    # order the folders
    folders = folders.sort_by(&:length)
    @jenkins_folders = folders
    logger.info "folders to create: #{folders.join(', ')}"
  end

  task :list_jobs => [
    :'terraform:prepare_environment',
    :'terraform:global_tfvars',
  ] do
    jobs = {
      "#{@terraform_environment}/terraform_code": {
        git_url: @terraform_global_tfvars['jenkins_terraform_git_url'],
        git_branch: "**/#{@terraform_global_tfvars['jenkins_terraform_git_branch']}",
        jenkinsfile: 'Jenkinsfile.terraform_code',
      },
      "#{@terraform_environment}/puppet_code": {
        git_url: @terraform_global_tfvars['jenkins_puppet_git_url'],
        git_branch: "**/#{@terraform_global_tfvars['jenkins_puppet_git_branch']}",
      },
    }

    # list terraform jobs
    Dir["tfvars/#{@terraform_environment}/*/*.tfvars"].each do |path|
      # only files
      next unless File.file?(path)

      # split and check for 3 segments
      parts = path.split('.tfvars').first.split('/')
      next if parts.length != 4
      stack = parts[-1]
      name = parts[-2]
      logger.debug "jobs to create: #{parts.inspect}"

      # env has to match
      jobs["#{@terraform_environment}/#{name}/terraform/#{stack}"] = {
        git_url: @terraform_global_tfvars['jenkins_terraform_git_url'],
        git_branch: "**/#{@terraform_global_tfvars['jenkins_terraform_git_branch']}",
        jenkinsfile: 'Jenkinsfile.terraform',
      }

    end

    # list packer jobs
    Dir['packer/*.json'].each do |path|
      # only files
      next unless File.file?(path)

      # split and check for 3 segments
      name = File.basename(path).split('.json').first
      jobs["#{@terraform_environment}/packer_#{name}"] = {
        git_url: @terraform_global_tfvars['jenkins_terraform_git_url'],
        git_branch: "**/#{@terraform_global_tfvars['jenkins_terraform_git_branch']}",
        jenkinsfile: 'Jenkinsfile.packer',
      }
    end

    @jenkins_jobs = jobs
    logger.info "jobs to create: #{jobs.keys.join(', ')}"
    logger.debug "jobs to create: #{jobs.inspect}"
  end

  task :initialize => [:list_folders, :list_jobs] do
    @jenkins = Jenkins.new(
      ENV['JENKINS_URL'],
      :user     => ENV['JENKINS_USER'],
      :password => ENV['JENKINS_PASSWORD'],
    )

    @jenkins_folders.each do |folder|
      @jenkins.create_folder folder
    end

    git_secret_name = "ssh_git_jenkins_#{@terraform_environment}"
    @jenkins_jobs.each do |job, opts|
      opts.update({
        git_secret: git_secret_name,
      })
      @jenkins.create_job(job.to_s, opts) 
    end

    @jenkins.create_ssh_credential(
      git_secret_name,
      'git',
      File.open('credentials/jenkins_key_pair').read
    )

    ssh_secret_name = "ssh_jenkins_#{@terraform_environment}"
    @jenkins.create_file_credential(
      ssh_secret_name,
      File.open('credentials/jenkins_key_pair').read
    )

    @jenkins.install_plugins('copyartifact@1.38.1','ansicolor@0.5.0')
    sleep 5
    @jenkins.safe_restart
  end
end
