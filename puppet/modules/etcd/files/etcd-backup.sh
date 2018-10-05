#!/bin/bash
set -euo pipefail

date=$(date -u +"%Y-%m-%d_%H-%M-%S")
backup_local_dir="/tmp/etcd-backup-${CLUSTER_NAME}"
backup_local_path="${backup_local_dir}/${date}-${CLUSTER_NAME}-snapshot.db"
mkdir -p ${backup_local_dir}

ETCDCTL_API=3 /opt/etcd-${ETCD_VERSION}/etcdctl --endpoints ${ENDPOINTS} --cert="${CA_PATH}.pem" --key="${CA_PATH}-key.pem" --cacert="${CA_PATH}-ca.pem" snapshot save ${backup_local_path}

backup_s3_path="s3://${BUCKET_NAME}/etcd-snapshot-${CLUSTER_NAME}/${date}-${CLUSTER_NAME}-snapshot.db"
aws configure set s3.signature_version s3v4
aws s3 cp --sse aws:kms ${backup_local_path} ${backup_s3_path}

rm -rf "${backup_local_path}"
