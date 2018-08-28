#!/bin/bash

set -euo pipefail

VAULT_UNSEALER_VERSION=0.1.4
VAULT_UNSEALER_HASH=7a01a119429b93edecb712aa897f2b22ba0575b7db5f810d4a9a40d993dad1aa
DEST_DIR=${DEST_DIR:-/opt/bin}

curl -sL https://github.com/jetstack/vault-unsealer/releases/download/${VAULT_UNSEALER_VERSION}/vault-unsealer_${VAULT_UNSEALER_VERSION}_linux_amd64 > ${DEST_DIR}/vault-unsealer

echo "${VAULT_UNSEALER_HASH}  ${DEST_DIR}/vault-unsealer" | sha256sum -c
chmod +x "${DEST_DIR}/vault-unsealer"
