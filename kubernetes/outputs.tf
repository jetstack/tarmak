output "kubectl_ssh_tunnel" {
  value = "ssh -N -L 8080:${aws_route53_record.kubernetes_master.fqdn}:8080 centos@ctm-bastion-${data.template_file.stack_name.rendered}"
}
