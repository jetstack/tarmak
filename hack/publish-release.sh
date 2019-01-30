#!/usr/bin/env bash
# Copyright 2017 The Kubernetes Authors.
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

# This command is used by bazel as the workspace_status_command
# to implement build stamping with git information.

set -o errexit
set -o nounset
set -o pipefail

export KUBE_ROOT=$(dirname "${BASH_SOURCE}")/..

source "${KUBE_ROOT}/hack/lib/version.sh"
kube::version::get_version_vars


# Make sure we are having the built binaries as first on the path with absolute
# path
export PATH=$(cd ${KUBE_ROOT}/bin; pwd):${PATH}

# Select sources from makefile or bazel
if [ -d "${KUBE_ROOT}/bazel-bin" ]; then
    SOURCES=(
    "${KUBE_ROOT}/bazel-bin/cmd/tarmak/linux_amd64_pure_stripped/tarmak_linux_amd64"
    "${KUBE_ROOT}/bazel-bin/cmd/tarmak/darwin_amd64_pure_stripped/tarmak_darwin_amd64"
    "${KUBE_ROOT}/bazel-bin/cmd/tagging_control/linux_amd64_pure_stripped/tagging_control_linux_amd64"
    "${KUBE_ROOT}/bazel-bin/cmd/wing/linux_amd64_pure_stripped/wing_linux_amd64"
    );
else
    SOURCES=(
    "${KUBE_ROOT}/_output/tarmak_linux_amd64"
    "${KUBE_ROOT}/_output/tarmak_darwin_amd64"
    "${KUBE_ROOT}/_output/tagging_control_linux_amd64"
    "${KUBE_ROOT}/_output/wing_linux_amd64"
    );
fi

mkdir -p "${KUBE_ROOT}/_release/"
WORK_DIR=`mktemp -d -p "${KUBE_ROOT}/_release/"`

# deletes the temp directory at exit
function cleanup {
  rm -rf "$WORK_DIR"
  echo "deleted temp working directory $WORK_DIR"
}
trap cleanup EXIT

# copy binaries to temp folder
dest_paths=()
for path in ${SOURCES[@]}; do
    dest_path=$(basename $path)
    dest_path=${dest_path//_linux/_${KUBE_GIT_VERSION}_linux}
    dest_path=${dest_path//_darwin/_${KUBE_GIT_VERSION}_darwin}
    dest_path="${WORK_DIR}/${dest_path}"
    cp -a $path "${dest_path}"
    chmod 0755 "${dest_path}"
    dest_paths+=($dest_path)
done

# compress binaries
upx -1 -q ${dest_paths[@]}

# generate hash
checksum_file=tarmak_${KUBE_GIT_VERSION}_checksums.txt
{
    cd "${WORK_DIR}"
    sha256sum * | tee "${checksum_file}"
}&
wait

# sign hashes
gpg -u tech+releases@jetstack.io --armor --output "${WORK_DIR}/${checksum_file}.asc"  --detach-sign "${WORK_DIR}/${checksum_file}"

# upload to github
ghr -u simonswine -r tarmak "${KUBE_GIT_VERSION}" "${WORK_DIR}"
