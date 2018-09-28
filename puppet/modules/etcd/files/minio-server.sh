#!/bin/bash
# Copyright Jetstack Ltd. See LICENSE for details.

set -euo pipefail

# download minio if neccessary
if [ ! -x /tmp/minio  ]; then
    curl -sL https://dl.minio.io/server/minio/release/linux-amd64/archive/minio.RELEASE.2018-09-25T21-34-43Z -o /tmp/minio.download
    echo "cc8d17c3384cbdf01f1f7a2390bd689f0821da1bd3371fc845a1822bab6bbc88  /tmp/minio.download" | sha256sum -c
    chmod +x /tmp/minio.download
    mv /tmp/minio.download /tmp/minio
fi
mkdir -p /tmp/minio-data/backup-bucket

if [ "${1:-}" == "download" ]; then
    exit 0
fi

exec /tmp/minio server /tmp/minio-data
