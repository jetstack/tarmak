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
          --input "tarmak/" \
          --input "cluster/"
          --clientset-path "github.com/jetstack/tarmak/pkg/client/" \
          --clientset-name internalclientset \
# Generate the versioned clientset (pkg/client/clientset_generated/clientset)
${BINDIR}/client-gen "$@" \
          --input-base "github.com/jetstack/tarmak/pkg/apis/" \
          --input "tarmak/v1alpha1" \
          --input "cluster/v1alpha1" \
          --clientset-path "github.com/jetstack/tarmak/pkg/" \
          --clientset-name "client" \
# generate lister
${BINDIR}/lister-gen "$@" \
          --input-dirs="github.com/jetstack/tarmak/pkg/apis/tarmak" \
          --input-dirs="github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1" \
          --input-dirs="github.com/jetstack/tarmak/pkg/apis/cluster" \
          --input-dirs="github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1" \
          --output-package "github.com/jetstack/tarmak/pkg/listers" \
# generate informer
${BINDIR}/informer-gen "$@" \
          --input-dirs "github.com/jetstack/tarmak/pkg/apis/tarmak" \
          --input-dirs "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1" \
          --input-dirs "github.com/jetstack/tarmak/pkg/apis/cluster" \
          --input-dirs "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1" \
          --internal-clientset-package "github.com/jetstack/tarmak/pkg/client/internalclientset" \
          --versioned-clientset-package "github.com/jetstack/tarmak/pkg/client" \
          --listers-package "github.com/jetstack/tarmak/pkg/listers" \
          --output-package "github.com/jetstack/tarmak/pkg/informers"
