#!/bin/bash

set -e

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_DEV_ROOT_TOKEN_ID=root-token
export VAULT_TOKEN=root-token


#Download vault binary
if [ ! -x /bin/vault  ]; then
    curl -sL -o /tmp/vault-dev.zip https://releases.hashicorp.com/vault/0.7.2/vault_0.7.2_linux_amd64.zip
    unzip /tmp/vault-dev.zip -d /tmp
    mv /tmp/vault /bin/vault
    chmod +x /bin/vault
    rm -f /tmp/vault-dev.zip
fi

#Download vault-helper binary
if [ ! -x /tmp/vault-helper  ]; then
    curl -sL https://github.com/jetstack-experimental/vault-helper/releases/download/0.8.2/vault-helper_0.8.2_linux_amd64 -o /tmp/vault-helper
    chmod +x /tmp/vault-helper
fi

mkdir -p /etc/vault
printf "init-token-client" > /etc/vault/init-token

exec /tmp/vault-helper dev-server test --init-token-etcd=init-token-etcd --init-token-master=init-token-master --init-token-worker=init-token-worker --init-token-all=init-token-client
