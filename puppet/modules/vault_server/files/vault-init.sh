#!/bin/bash

/opt/bin/vault status

if [ $? -eq 0 ]; then
  exit 0
else
  /opt/bin/vault-unsealer init --overwrite-existing --init-root-token dev-root-token --store-root-token
fi
