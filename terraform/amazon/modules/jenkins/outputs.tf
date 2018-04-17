output "jenkins_fqdn" {
  value = "${aws_route53_record.jenkins.fqdn}"
}

output "jenkins_security_group_id" {
  value = "${aws_security_group.jenkins.id}"
}

output "jenkins_dns_name" {
  value = "${aws_elb.jenkins.dns_name}"
}

output "jenkins_url" {
  value = "https://${aws_route53_record.jenkins_elb.fqdn}"
}