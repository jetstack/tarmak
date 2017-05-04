output "stack_name" {
  value = "${data.template_file.stack_name.rendered}"
}

output "environment" {
  value = "${var.environment}"
}

output "public_zone_id" {
  value = "${aws_route53_zone.public.id}"
}

output "public_zone" {
  value = "${aws_route53_zone.public.name}"
}

output "public_zone_name_servers" {
  value = "${aws_route53_zone.public.name_servers}"
}

output "jenkins_data_volume_id" {
  value = "${aws_ebs_volume.jenkins.id}"
}

output "puppet_master_data_volume_id" {
  value = "${aws_ebs_volume.puppet_master.id}"
}

output "bucket_prefix" {
  value = "${var.bucket_prefix}"
}
