require 'spec_helper_acceptance'

describe '::vault_server' do
  version = '0.9.5'

  before(:all) do
    hosts.each do |host|
      if fact_on(host, 'osfamily') == 'RedHat'
        on(host, 'yum install -y unzip')
      elsif fact_on(host, 'osfamily') == 'Debian'
        on(host, 'apt-get install -y unzip iptables')
      end

      on host, "iptables -F INPUT"

      on host, 'ln -sf /etc/puppetlabs/code/modules/vault_server/files/vault-dev-server.service /etc/systemd/system/vault-dev-server.service'
      on host, 'systemctl daemon-reload'
      on host, 'systemctl start vault-dev-server.service'
    end
  end

  context 'vault with non refreshable token' do
    # Using puppet_apply as a helper
    it 'should work with no errors based on the example' do
      pp = <<-EOS
class {'vault_client':
  version => '#{version}',
  init_token => 'init-token-all',
  init_role => 'test-all',
}
EOS
      # cleanup existing config
      shell('rm -rf /etc/vault/init-token /etc/vault/token')

      # Run it twice and test for idempotency
      apply_manifest(pp, :catch_failures => true)
      expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero
    end

    it 'runs the correct version of vault-helper' do
      show_result = shell('/opt/bin/vault-helper version')
      expect(show_result.stdout).to match(/#{version}/)
    end

    it 'runs renew-token without error' do
        result = shell('export VAULT_ADDR=http://127.0.0.1:8200; export VAULT_INIT_ROLE=test-all; /opt/bin/vault-helper renew-token')
      expect(result.exit_code).to eq(0)
    end

    it 'requests a client cert from test-ca' do
      pp = <<-EOS
class {'vault_client':
  version => '#{version}',
  init_token => 'init-token-all',
  init_role => 'test-all',
}

vault_client::cert_service{ 'test-client':
  common_name  => 'test-client',
  base_path    => '/tmp/test-cert-client',
  role         => 'test/pki/k8s/sign/kube-apiserver',
}
EOS
      apply_manifest(pp, :catch_failures => true)
      expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero

      # check CN
      result = shell('openssl x509 -noout -subject -in /tmp/test-cert-client.pem')
      expect(result.exit_code).to eq(0)
      expect(result.stdout).to match(/CN=test-client$/)
    end

    it 'requests new cert for a changed common_name' do
      pp = <<-EOS
class {'vault_client':
  version => '#{version}',
  init_token => 'init-token-all',
  init_role => 'test-all',
}

vault_client::cert_service{ 'test-client':
  common_name  => 'test-client-aa',
  base_path    => '/tmp/test-cert-client',
  role         => 'test/pki/k8s/sign/kube-apiserver',
}
EOS
      apply_manifest(pp, :catch_failures => true)
      expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero

      result = shell('openssl x509 -noout -subject -in /tmp/test-cert-client.pem')
      expect(result.exit_code).to eq(0)
      expect(result.stdout).to match(/CN=test-client-aa$/)
    end

    it 'requests new cert for a added IP/DNS SANs' do
      pp = <<-EOS
class {'vault_client':
  version => '#{version}',
  init_token => 'init-token-all',
  init_role => 'test-all',
}

vault_client::cert_service{ 'test-client':
  common_name  => 'test-client-aa',
  base_path    => '/tmp/test-cert-client',
  role         => 'test/pki/k8s/sign/kube-apiserver',
  ip_sans      => ['8.8.4.4','8.8.8.8'],
  alt_names    => ['public-dns-4.google','public-dns-8.google'],
}
EOS
      apply_manifest(pp, :catch_failures => true)
      expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero

      result = shell('openssl x509 -noout -text -in /tmp/test-cert-client.pem')
      expect(result.exit_code).to eq(0)
      expect(result.stdout).to match(/CN=test-client-aa/)
      expect(result.stdout).to match(/DNS:public-dns-4\.google/)
      expect(result.stdout).to match(/DNS:public-dns-8\.google/)
      expect(result.stdout).to match(/IP Address:8\.8\.4\.4/)
      expect(result.stdout).to match(/IP Address:8\.8\.8\.8/)
    end
  end

  context 'vault with init_token' do
    # Using puppet_apply as a helper
    it 'should work with no errors based on the example' do
      pp = <<-EOS
class {'vault_client':
  version => '#{version}',
  init_token => 'init-token-all',
  init_role => 'test-all',
}
EOS
      # cleanup existing config
      shell('rm -rf /etc/vault/init-token /etc/vault/token')

      # Run it twice and test for idempotency
      apply_manifest(pp, :catch_failures => true)
      expect(apply_manifest(pp, :catch_failures => true).exit_code).to be_zero
    end

    it 'renews tokens without error' do
      result = shell('systemctl start vault-token-renewal.service')
      expect(result.exit_code).to eq(0)
    end
  end
end
