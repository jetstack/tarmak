#!/bin/bash

set -ex

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root-token-dev

vault read -format=json test/pki/k8s/roles/kubelet | python -c "import socket, sys, json; v=json.load(sys.stdin); v=v['data']; k='allowed_domains'; d=v[k]; d.append(socket.gethostname()); v[k] = ','.join(list(set(d))); print json.dumps(v)" | vault write test/pki/k8s/roles/kubelet -
