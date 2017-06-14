require 'puppetlabs_spec_helper/module_spec_helper'
require 'open3'
require 'logger'

$logger = Logger.new(STDERR)
$logger.level = Logger::DEBUG

module MinikubeHelpers
  def help
    :available
  end

  def cmd(*cmd)
    Open3.popen3(*cmd) do |stdin, stdout, stderr, wait_thr|
      stdin.close
      $logger.debug "cmd=#{cmd} stdout=#{stdout.read} stderr=#{stderr.read} exitcode=#{wait_thr.value.exitstatus}"
      wait_thr.value.exitstatus
    end
  end

  def cmd_stdout(*cmd)
    Open3.popen3(*cmd) do |stdin, stdout, stderr, wait_thr|
      stdin.close
      $logger.debug "cmd=#{cmd} stderr=#{stderr.read} exitcode=#{wait_thr.value.exitstatus}"
      stdout.read
    end
  end

  def minikube_cmd(subcmd)
    return ['minikube', subcmd, '--profile', @minikube_profile]
  end

  def minikube_start
    cmd(
      *minikube_cmd('start'),
      '--extra-config=apiserver.Authorization.Mode=RBAC',
      "--kubernetes-version=v#{@minikube_kubernetes_version}",
    )
  end

  def minikube_status
    cmd(*minikube_cmd('status'))
  end

  def kubectl_apply(manifests)
    kubectl_stdin('apply', manifests)
  end

  def kubectl_delete(manifests)
    kubectl_stdin('delete', manifests)
  end

  def kubectl_create(manifests)
    kubectl_stdin('create', manifests)
  end

  def kubectl_stdin(subcmd, manifests)
    cmd = ['kubectl', subcmd, '-f', '-']
    Open3.popen3(*cmd) do |stdin, stdout, stderr, wait_thr|
      manifests.each do |manifest|
        stdin.write("---\n")
        stdin.write(manifest)
      end
      stdin.close
      $logger.debug "cmd=#{cmd} stdout=#{stdout.read} stderr=#{stderr.read} exitcode=#{wait_thr.value.exitstatus}"
      return wait_thr.value.exitstatus
    end
  end

  def kubectl_get(object, namespace, name)
    return JSON.parse cmd_stdout(
      'kubectl', 'get', object, '--namespace', namespace, '-o', 'json', name
    )
  end

  def minikube_wait_deployment_ready(namespace, name, replicas)
    status = JSON.parse kubectl_get('deployment', namespace, name)
    return status
  end

  def minikube_prepare
    @minikube_profile = "kubernetes-addons"
    @minikube_kubernetes_version = ENV['KUBERNETES_VERSION'] || '1.6.4'
    ENV['KUBECONFIG'] = "#{Dir.pwd}/kubeconfig"
    if minikube_status != 0
      minikube_start
    end
  end

  def minikube_apply(manifests)
    minikube_prepare
    kubectl_apply(manifests)
  end
end

RSpec.configure do |config|
  config.default_facts = {
    :path => '/bin:/sbin:/usr/bin:/usr/sbin:/opt/bin',
    :osfamily => 'RedHat',
  }
  config.include MinikubeHelpers

end
