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
    stdout = cmd_stdout(*minikube_cmd('status'))
    $logger.debug "minikube status  stdout=#{stdout}"
    return true if stdout.scan(/Running/).count == 2
    false
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

  def minikube_fix_kube_dns
    cmd('kubectl', 'create', 'serviceaccount', '--namespace', 'kube-system', 'kube-dns')
    fix_job_yaml = <<-EOS
apiVersion: batch/v1
kind: Job
metadata:
  name: fix-kube-dns
spec:
  activeDeadlineSeconds: 300
  template:
    metadata:
      name: fix-kube-dns
    spec:
      containers:
      - name: fix-kube-dns
        image: ruby:alpine
        command: [
        "ruby",
        "-r", "yaml",
        "-e",
        "\
        path = '/ETC/kubernetes/addons/kube-dns-controller.yaml';\
        d = YAML.load_file(path);\
        d['spec']['template']['spec']['serviceAccountName'] = 'kube-dns';\
        File.open(path, 'w') {|f| f << d.to_yaml }\
        "
        ]
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /ETC
          name: host-etc
      volumes:
      - name: host-etc
        hostPath:
          path: /etc
      restartPolicy: Never
EOS
    kubectl_apply([fix_job_yaml])
  end

  def minikube_wait_deployment_ready(namespace, name, replicas)
    status = JSON.parse kubectl_get('deployment', namespace, name)
    return status
  end

  def minikube_prepare
    @minikube_profile = "kubernetes-addons"
    @minikube_kubernetes_version = ENV['KUBERNETES_VERSION'] || '1.6.4'
    ENV['KUBECONFIG'] = "#{Dir.pwd}/kubeconfig"
    if ! minikube_status
      minikube_start
    end

    # check kube-dns healthiness
    retries = 10
    while true do
      begin
        ready_replicas = kubectl_get('deployment','kube-system', 'kube-dns')['status']['readyReplicas']
        break if not ready_replicas.nil? and ready_replicas > 0
      rescue Exception => e
        $logger.warn "kube-dns: #{e}"
      end
      raise "kube-dns is not ready" if retries == 0
      $logger.debug "kube-dns not ready yet, applying updates to support RBAC and retry"
      minikube_fix_kube_dns
      retries -= 1
      sleep 10
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
