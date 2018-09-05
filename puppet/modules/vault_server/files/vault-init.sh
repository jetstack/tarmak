#!/bin/bash

if [ -e "/etc/vault/vault-shared-key" ]; then
  cp /etc/vault/vault-shared-key /etc/vault/vault-unseal-0
  exit 0
fi

/opt/bin/vault status

EXITCODE=$?

if [ $EXITCODE -eq 0 ]
then
  cat /etc/vault/vault-unseal-0 | base64 | cat > /etc/vault/vault-shared-key
  scp -o StrictHostKeyChecking=no /etc/vault/vault-shared-key root@vault-1:/etc/vault/vault-shared-key
  scp -o StrictHostKeyChecking=no /etc/vault/vault-shared-key root@vault-2:/etc/vault/vault-shared-key
  scp -o StrictHostKeyChecking=no /etc/vault/vault-shared-key root@vault-3:/etc/vault/vault-shared-key

  exit 1
else
    /opt/bin/vault-unsealer init --overwrite-existing --init-root-token dev-root-token

  exit 1
fi
