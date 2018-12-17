#!/bin/bash

# Copyright 2018 The Jetstack Tarmak contributors.
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

set -e

REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY:-"$(cd "$(dirname "$0")" && pwd -P)"/..}
cd ${REPO_ROOT}
export PATH=$PATH:${REPO_ROOT}/bin/

GIT_TAG=$(git describe --tags --abbrev=0)

CMD_PATH="docs/generated/cmd"
CMD_ROOT=$(cd "${REPO_ROOT}/${CMD_PATH}" 2> /dev/null && pwd -P)
OUTPUT_DIR="${CMD_ROOT}"

BINDIR=$REPO_ROOT/bin
HACKDIR=$REPO_ROOT/hack

echo "+++ Building cmd-gen"
go build -o ${BINDIR}/cmd-gen ./hack/cmd-gen

echo "+++ Removing old output"
rm -Rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

echo "+++ Running cmd-gen"
${BINDIR}/cmd-gen ${OUTPUT_DIR}

echo "+++ referencing output in docs"
cat > "${REPO_ROOT}/docs/cmd-docs.rst" << EOF
==========================
Command Line Documentation
==========================

Command line documentation for both tarmak and wing commands

EOF

for cmd in "tarmak" "wing"; do
    farray=$(basename -s .rst -a ${OUTPUT_DIR}/${cmd}/* | sort)
    for f in ${farray}; do
        cat >> "${REPO_ROOT}/docs/cmd-docs.rst" << EOF
.. toctree::
   :maxdepth: 1

   generated/cmd/${cmd}/${f}

EOF
    done
done

echo "+++ Command docs created"
