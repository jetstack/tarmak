#!/bin/bash

set -ex

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root-token-dev

#Download jq binary
if [ ! -x /bin/jq ]; then
    curl -sL -o jq https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64
    mv jq /bin/jq
    chmod +x /bin/jq
fi

vault read -format=json test/pki/k8s/roles/kubelet | jq ".data | .allowed_domains += [\"$(hostname)\"]" | vault write test/pki/k8s/roles/kubelet -
