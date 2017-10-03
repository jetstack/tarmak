#!/bin/bash

# The only argument this script should ever be called with is '--verify-only'

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE}")/..
BINDIR=${REPO_ROOT}/bin

# Generate the internal clientset (pkg/client/clientset_generated/internalclientset)
${BINDIR}/client-gen "$@" \
          --input-base "github.com/jetstack/tarmak/pkg/apis/" \
          --input "wing" \
          --clientset-path "github.com/jetstack/tarmak/pkg/wing/clients" \
          --clientset-name internalclientset \
# Generate the versioned clientset (pkg/client/clientset_generated/clientset)
${BINDIR}/client-gen "$@" \
          --input-base "github.com/jetstack/tarmak/pkg/apis/" \
          --input "wing/v1alpha1" \
          --clientset-path "github.com/jetstack/tarmak/pkg/wing" \
          --clientset-name "client" \
# generate lister
${BINDIR}/lister-gen "$@" \
          --input-dirs="github.com/jetstack/tarmak/pkg/apis/wing" \
          --input-dirs="github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1" \
          --output-package "github.com/jetstack/tarmak/pkg/wing/listers" \
# generate informer
${BINDIR}/informer-gen "$@" \
          --input-dirs="github.com/jetstack/tarmak/pkg/apis/wing" \
          --input-dirs="github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1" \
          --internal-clientset-package "github.com/jetstack/tarmak/pkg/wing/clients/internalclientset" \
          --versioned-clientset-package "github.com/jetstack/tarmak/pkg/wing/client" \
          --listers-package "github.com/jetstack/tarmak/pkg/wing/listers" \
          --output-package "github.com/jetstack/tarmak/pkg/wing/informers"
