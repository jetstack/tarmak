#!/bin/bash
set -euo pipefail

DATE=$(date -u +"%Y-%m-%d_%H-%M-%S")
backup_local_path=/tmp/etcd-backup-${DATE}
backup_local_events_path="${backup_local_path}/k8s-events-snapshot.db"
backup_local_main_path="${backup_local_path}/k8s-main-snapshot.db"
backup_local_overlay_path="${backup_local_path}/overlay-snapshot.db"
mkdir -p ${backup_local_path}

ETCDCTL_API=3 /opt/bin/etcdctl --endpoints ${K8S_EVENTS_ENDPOINTS} --cert="${K8S_EVENTS_CA}.pem" --key="${K8S_EVENTS_CA}-key.pem" --cacert="${K8S_EVENTS_CA}-ca.pem" snapshot save ${backup_local_events_path}
ETCDCTL_API=3 /opt/bin/etcdctl --endpoints ${K8S_MAIN_ENDPOINTS} --cert="${K8S_MAIN_CA}.pem" --key="${K8S_MAIN_CA}-key.pem" --cacert="${K8S_MAIN_CA}-ca.pem"  snapshot save ${backup_local_main_path}
ETCDCTL_API=3 /opt/bin/etcdctl --endpoints ${OVERLAY_ENDPOINTS} --cert="${OVERLAY_CA}.pem" --key="${OVERLAY_CA}-key.pem" --cacert="${OVERLAY_CA}-ca.pem" snapshot save ${backup_local_overlay_path}

backup_s3_prefix="s3://${BUCKET_NAME}/etcd-snapshot-${DATE}"
backup_s3_events_path="${backup_s3_prefix}/k8s-events-snapshot.db"
backup_s3_main_path="${backup_s3_prefix}/k8s-main-snapshot.db"
backup_s3_overlay_path="${backup_s3_prefix}/overlay-snapshot.db"
aws configure set s3.signature_version s3v4
aws s3 cp --sse aws:kms ${backup_local_events_path} ${backup_s3_events_path}
aws s3 cp --sse aws:kms ${backup_local_main_path} ${backup_s3_main_path}
aws s3 cp --sse aws:kms ${backup_local_overlay_path} ${backup_s3_overlay_path}

rm -rf "${backup_local_path}"
