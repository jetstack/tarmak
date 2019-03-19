class kubernetes::aws_iam_authenticator_init(
  String $auth_token_webhook_file,
  Array[String] $systemd_wants = [],
  Array[String] $systemd_requires = [],
  Array[String] $systemd_after = [],
  Array[String] $systemd_before = [],
){
  require ::kubernetes

  $service_name = 'aws-iam-authenticator-init'

  $_systemd_wants = $systemd_wants
  $_systemd_after = $systemd_after
  $_systemd_requires = $systemd_after
  $_systemd_before = $systemd_before

  file{"${::kubernetes::systemd_dir}/${service_name}.service":
    ensure  => file,
    mode    => '0644',
    owner   => 'root',
    group   => 'root',
    content => template("kubernetes/${service_name}.service.erb"),
  }
  ~> exec { "${service_name}-daemon-reload":
    command     => 'systemctl daemon-reload',
    path        => $::kubernetes::path,
    refreshonly => true,
  }
  -> service{ "${service_name}.service":
    enable  => true,
  }
}
