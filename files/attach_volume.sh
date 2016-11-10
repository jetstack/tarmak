#!/bin/bash
REGION=$(curl -s http://169.254.169.254/latest/dynamic/instance-identity/document | grep region | awk -F\" '{print $4}')
INSTANCE_ID=$(curl -s http://169.254.169.254/latest/dynamic/instance-identity/document | grep instanceId | awk -F\" '{print $4}')
KEY="Etcd_Volume_Attach"
VOLUME_NAME=$(aws ec2 describe-tags --filters "Name=resource-id,Values=$INSTANCE_ID" "Name=key,Values=$KEY" --region=$REGION --output=text | cut -f5)
VOLUME_ID=$(aws ec2 describe-volumes --region=$REGION --filters "Name=tag:Name,Values=$VOLUME_NAME" --query "Volumes[0].VolumeId" --output text)
VOLUME_ATTACHED=$(aws ec2 describe-volumes --region=$REGION --filters "Name=tag:Name,Values=$VOLUME_NAME" --query "Volumes[0].Attachments" --output text | grep /dev/xvdd)

if [ "$VOLUME_ATTACHED" == "" ]; then
  aws ec2 attach-volume --volume-id $VOLUME_ID --instance-id $INSTANCE_ID --region=$REGION --device /dev/xvdd
fi

until [ "$VOLUME_STATUS_ATTACHED" != "" ]
  do
    VOLUME_STATUS_ATTACHED=$(aws ec2 describe-volumes --region=$REGION --filters "Name=tag:Name,Values=$VOLUME_NAME" --query "Volumes[0].Attachments" --output text | grep /dev/xvdd | grep -i attached)
      sleep 2
done

