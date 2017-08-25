#!/bin/bash

set -e

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_DEV_ROOT_TOKEN_ID=init-token-client
export VAULT_TOKEN=init-token-client


#Download vault binary
if [ ! -x /bin/vault  ]; then
    curl -sL -o /tmp/vault-dev.zip https://releases.hashicorp.com/vault/0.7.2/vault_0.7.2_linux_amd64.zip
    unzip /tmp/vault-dev.zip -d /tmp
    mv /tmp/vault /bin/vault
    chmod +x /bin/vault
    rm -f /tmp/vault-dev.zip
fi

printf "init-token-client" > /etc/vault/init-token


exec /opt/bin/vault-helper dev-server test --init-token-etcd=init-token-etcd --init-token-master=init-token-master --init-token-worker=init-token-worker --init-token-all=init-token-client
