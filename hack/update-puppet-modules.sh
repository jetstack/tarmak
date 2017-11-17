#!/bin/bash

# The only argument this script should ever be called with is '--verify-only'

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE}")/..

PUPPET_MODULES="aws_ebs calico etcd kubernetes kubernetes_addons prometheus tarmak vault_client"

cd "${REPO_ROOT}"

for module_name in ${PUPPET_MODULES}; do
    git subtree pull --prefix "puppet/modules/${module_name}" "git@github.com:jetstack/puppet-module-${module_name}.git" master
done
