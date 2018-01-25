output "stack_name" {
  value = "${data.template_file.stack_name.rendered}"
}

output "environment" {
  value = "${var.environment}"
}

output "public_zone_id" {
  value = "${var.public_zone_id}"
}

output "public_zone" {
  value = "${var.public_zone}"
}

output "jenkins_data_volume_id" {
  value = "${aws_ebs_volume.jenkins.*.id}"
}

output "bucket_prefix" {
  value = "${var.bucket_prefix}"
}
