#!/bin/bash

set -euo pipefail

backup_local_path=/tmp/consul-backup
backup_local_acls_path=/tmp/consul-backup-acls
/opt/bin/consul-backinator backup -file "${backup_local_path}" -acls "${backup_local_acls_path}"
backup_s3_prefix="s3://${BUCKET_NAME}/consul-backup-$(date -u +"%Y-%m-%d_%H-%M-%S")"
backup_s3_path="${backup_s3_prefix}/consul-backup"
backup_s3_acls_path="${backup_s3_prefix}/consul-backup-acls"
aws s3 cp --sse aws:kms "${backup_local_path}"          "${backup_s3_path}"
aws s3 cp --sse aws:kms "${backup_local_path}.sig"      "${backup_s3_path}.sig"
aws s3 cp --sse aws:kms "${backup_local_acls_path}"     "${backup_s3_acls_path}"
aws s3 cp --sse aws:kms "${backup_local_acls_path}.sig" "${backup_s3_acls_path}.sig"
rm -rf "${backup_local_path}" "${backup_local_path}.sig" "${backup_local_acls_path}" "${backup_local_acls_path}.sig"
