require 'logger'
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

  desc 'login using multi-factor-auth'
  task :login_mfa => :prepare do

    logger.debug "credentials: #{@aws_credentials}"
    logger.debug "config: #{@aws_config}"
    duration = 86400

    if @aws_config['mfa_serial'].nil?
      fail "No mfa_serial in AWS config #{@aws_config_file} for profile #{@aws_profile}"
    end

    if ENV['MFA_TOKEN']
      token = ENV['MFA_TOKEN']
    else
      require 'highline'
      hl = HighLine.new($stdin, $stderr)
      token = hl.ask('Enter MFA token: ')
    end

    logger.info "generate temporary credentials using aws profile '#{@aws_profile}', mfa #{@aws_config['mfa_serial']}, token '#{token}'"
    sts = Aws::STS::Client.new(credentials: @aws_credentials)
    credentials =sts.get_session_token(
      :serial_number => @aws_config['mfa_serial'],
      :token_code => token,
      :duration_seconds => duration,
    ).credentials

    puts "export AWS_ACCESS_KEY_ID=#{credentials.access_key_id}"
    puts "export AWS_SECRET_ACCESS_KEY=#{credentials.secret_access_key}"
    puts "export AWS_SESSION_TOKEN=#{credentials.session_token}"
  end
end

namespace :terraform do
  task :prepare => :'aws:prepare' do

    @terraform_names = /^[a-z0-9]{3,16}$/
    if not @terraform_names.match(ENV['TERRAFORM_NAME'])
      fail "Please provide a TERRAFORM_NAME variable with that matches #{@terraform_names}"
    end
    @terraform_name = ENV['TERRAFORM_NAME']

    @terraform_environments = ['nonprod']
    @terraform_environment = ENV['TERRAFORM_ENVIRONMENT'] || 'nonprod'
    unless @terraform_environments.include?(@terraform_environment)
      fail "Please provide a TERRAFORM_ENVIRONMENT out of #{@terraform_environments}"
    end

    @terraform_stacks = ['network', 'vault', 'tools', 'kubernetes']
    unless @terraform_stacks.include?(ENV['TERRAFORM_STACK'])
      fail "Please provide a TERRAFORM_STACK out of #{@terraform_stacks}"
    end
    @terraform_stack = ENV['TERRAFORM_STACK']

    sh 'mkdir -p tfstate'
    @terraform_state_bucket = "skyscanner-k8s-#{@terraform_environment}-#{@aws_region}-terraform-state"
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

  task :plan => :prepare do
    Dir.chdir(@terraform_stack) do
      sh 'terraform', 'plan', *@terraform_args
    end
  end

  task :apply => :prepare do
    Dir.chdir(@terraform_stack) do
      sh 'terraform', 'apply', *@terraform_args
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
