resource "aws_cloudwatch_metric_alarm" "autorecover_etcd" {
  count               = "${var.etcd_instance_count}"
  alarm_name          = "kubernetes-${data.template_file.stack_name.rendered}-etcd-autorecover-${count.index}"
  namespace           = "AWS/EC2"
  evaluation_periods  = "2"
  period              = "60"
  alarm_description   = "This metric auto recovers EC2 instances"
  alarm_actions       = ["arn:aws:automate:${var.region}:ec2:recover"]
  statistic           = "Minimum"
  comparison_operator = "GreaterThanThreshold"
  threshold           = "1"
  metric_name         = "StatusCheckFailed_System"

  dimensions {
    InstanceId = "${element(aws_instance.etcd.*.id, count.index)}"
  }
}
