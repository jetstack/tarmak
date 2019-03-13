#!/bin/sh

set -e
set -x

CONSUL_DATACENTER=$(strings -n1 $1/data/raft/raft.db | grep Datacenter -A1 | head -n2 | tail -n 1)

if [ -z "${CONSUL_DATACENTER}" ]; then
  echo "{}"
else
  echo "{\"datacenter\":\"${CONSUL_DATACENTER}\",\"acl_datacenter\":\"${CONSUL_DATACENTER}\"}"
fi
