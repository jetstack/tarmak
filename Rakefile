require 'logger'
logger = Logger.new(STDERR)
logger.level = Logger::DEBUG

namespace :aws do
  task :prepare do
    require 'aws-sdk'
    require 'inifile'
    @aws_profile = ENV['AWS_PROFILE'] || 'ss_non_prod'
    @aws_config_file = IniFile.load(ENV['HOME'] + '/.aws/config')
    @aws_config = @aws_config_file["profile #{@aws_profile}"]
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
