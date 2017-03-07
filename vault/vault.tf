data "template_file" "vault" {
  template = "${file("${path.module}/templates/vault_user_data.yaml")}"
  count    = "${var.instance_count}"

  vars {
    FQDN           = "vault-${count.index + 1}.${data.terraform_remote_state.network.private_zones[0]}"
    ENVIRONMENT    = "${var.environment}"
    REGION         = "${var.region}"
    INSTANCE_COUNT = "${var.instance_count}"
    VAULT_VERSION  = "${var.vault_version}"
    VOLUME_ID      = "${element(aws_ebs_volume.vault.*.id, count.index)}"

    CONSUL_VERSION      = "${var.consul_version}"
    CONSUL_MASTER_TOKEN = "${var.consul_master_token}"
    CONSUL_ENCRYPT      = "${var.consul_encrypt}"

    VAULT_TLS_CERT_PATH = "s3://${data.terraform_remote_state.network.secrets_bucket}/vault-${var.environment}/cert.pem"
    VAULT_TLS_KEY_PATH  = "s3://${data.terraform_remote_state.network.secrets_bucket}/vault-${var.environment}/cert-key.pem"
    VAULT_TLS_CA_PATH   = "s3://${data.terraform_remote_state.network.secrets_bucket}/vault-${var.environment}/ca.pem"

    BUCKET_BACKUP = "${aws_s3_bucket.vault-backup.bucket}"

    # run backup once per instance spread throughout the day
    BACKUP_SCHEDULE = "*-*-* ${format("%02d",count.index * (24/var.instance_count))}:00:00"
  }
}

resource "aws_cloudwatch_metric_alarm" "vault-autorecover" {
  count               = "${var.instance_count}"
  alarm_name          = "vault-autorecover-${var.environment}-${count.index+1}"
  namespace           = "AWS/EC2"
  evaluation_periods  = "2"
  period              = "60"
  alarm_description   = "This metric auto recovers Vault instances for the ${var.environment} cluster"
  alarm_actions       = ["arn:aws:automate:${var.region}:ec2:recover"]
  statistic           = "Minimum"
  comparison_operator = "GreaterThanThreshold"
  threshold           = "1"
  metric_name         = "StatusCheckFailed_System"

  dimensions {
    InstanceId = "${element(aws_instance.vault.*.id, count.index)}"
  }
}

resource "aws_instance" "vault" {
  ami                  = "${var.coreos_ami[var.region]}"
  instance_type        = "${var.vault_instance_type}"
  key_name             = "${var.key_name}"
  subnet_id            = "${element(data.terraform_remote_state.network.private_subnet_ids, count.index % length(data.terraform_remote_state.network.availability_zones))}"
  count                = "${var.instance_count}"
  user_data            = "${element(data.template_file.vault.*.rendered, count.index)}"
  iam_instance_profile = "${element(aws_iam_instance_profile.vault.*.name, count.index)}"
  private_ip           = "${cidrhost(element(data.terraform_remote_state.network.private_subnets, count.index % length(data.terraform_remote_state.network.availability_zones)),(10 + (count.index/length(data.terraform_remote_state.network.availability_zones))))}"

  vpc_security_group_ids = [
    "${aws_security_group.vault.id}",
  ]

  tags {
    Name        = "${data.template_file.stack_name.rendered}-vault-${count.index+1}"
    Environment = "${var.environment}"
    Project     = "${var.project}"
    Contact     = "${var.contact}"
    VaultCluster = "${var.environment}"
  }
}

resource "aws_ebs_volume" "vault" {
  count             = "${var.instance_count}"
  size              = "${var.vault_data_size}"
  availability_zone = "${element(data.terraform_remote_state.network.availability_zones, count.index % length(data.terraform_remote_state.network.availability_zones))}"

  tags {
    Name         = "${data.template_file.stack_name.rendered}-vault-${count.index+1}"
    Environment  = "${var.environment}"
    Project      = "${var.project}"
    Contact      = "${var.contact}"
  }

  lifecycle = {
    #prevent_destroy = true
  }
}
