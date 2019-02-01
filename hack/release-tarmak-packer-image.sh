#!/usr/bin/env bash

# Copyright 2018 The Jetstack tarmak contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

export KUBE_ROOT=$(dirname "${BASH_SOURCE}")/..

source "${KUBE_ROOT}/hack/lib/version.sh"
kube::version::get_version_vars

TARMAK_BASE_IMAGE_NAME=${TARMAK_BASE_IMAGE_NAME:-centos-puppet-agent}
TARMAK_VERSION=${TARMAK_VERSION:-${KUBE_GIT_VERSION}}

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..
PACKER_ROOT=$(dirname "${BASH_SOURCE}")/../packer/amazon

TMP_PACKER_CONFIG="${PACKER_ROOT}/.tmp_${TARMAK_BASE_IMAGE}.json"

# build packer release config
jq -s '
  .[2] = (
    .[0]
    | del(
    .variables.tarmak_environment,
    .variables.ebs_volume_encrypted,
    .builders[0].source_ami_filter,
    .builders[0].tags.tarmak_environment)
    )
  | .[2].variables = .[2].variables * .[1].variables
  | .[2].builders[0] = .[2].builders[0] * .[1].builders[0]
  | .[2]' "${PACKER_ROOT}/${TARMAK_BASE_IMAGE_NAME}.json" "${PACKER_ROOT}/releases.json" > "${TMP_PACKER_CONFIG}"

# run packer
export TARMAK_BASE_IMAGE TARMAK_VERSION
exec packer build "${TMP_PACKER_CONFIG}"
