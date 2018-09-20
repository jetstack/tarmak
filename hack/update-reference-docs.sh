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

GIT_TAG=$(git describe --tags --abbrev=0)

REFERENCE_PATH="docs/generated/reference"
REFERENCE_ROOT=$(cd "${REPO_ROOT}/${REFERENCE_PATH}" 2> /dev/null && pwd -P)
OUTPUT_DIR="${REFERENCE_ROOT}/includes"

BINDIR=$REPO_ROOT/bin
HACKDIR=$REPO_ROOT/hack

cleanup() {
    pushd "${REFERENCE_ROOT}"
    echo "+++ Cleaning up temporary docsgen files"
    # Clean up old temporary files
    rm -Rf "openapi-spec" "includes" "manifest.json" "openapi" "static_includes" "brodocs"
    popd
}

# Ensure we start with a clean set of directories
trap cleanup EXIT
cleanup

echo "+++ Removing old output"
rm -Rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

echo "+++ Running openapi-gen"
${BINDIR}/openapi-gen \
        --input-dirs github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1,github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1,github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1\
        --output-package "github.com/jetstack/tarmak/${REFERENCE_PATH}/openapi"\
        --go-header-file "${HACKDIR}/boilerplate/boilerplate.go.txt"

## Generate swagger.json from the Golang generated openapi spec
echo "+++ Running 'swagger-gen' to generate swagger.json"
mkdir -p "${REFERENCE_ROOT}/openapi-spec" "${REFERENCE_ROOT}/openapi-spec"
go build -o ${HACKDIR}/swagger-gen/swagger-gen ${HACKDIR}/swagger-gen/main.go
${HACKDIR}/swagger-gen/swagger-gen

echo "+++ Running gen-apidocs"
mkdir -p ${REFERENCE_ROOT}/static_includes
$BINDIR/gen-apidocs \
    --copyright "<a href=\"https://jetstack.io\">Copyright 2018 Jetstack Ltd.</a>" \
    --title "Tarmak API Reference" \
    --config-dir "${REFERENCE_ROOT}"

echo "+++ Running brodocs"
OUTPUT_DIR="${REFERENCE_ROOT}/output"
BRODOC_DIR="${REFERENCE_ROOT}/brodocs"
BRODOC_VEN=./vendor/github.com/Birdrock/brodocs
rm -rf ${OUTPUT_DIR}
cp -r ${BRODOC_VEN} ${REFERENCE_ROOT}/.
cp ${REFERENCE_ROOT}/manifest.json ${BRODOC_DIR}/.
rm -rf ${BRODOC_DIR}/documents/* && cp -r ${REFERENCE_ROOT}/includes/* ${BRODOC_DIR}/documents/
cd ${BRODOC_DIR} && npm update && npm install && node brodoc.js && cd ../.
mkdir -p ${OUTPUT_DIR} && cp -r ${BRODOC_DIR}/{*.js,*.html,*.css,documents/*,node_modules} ${OUTPUT_DIR}/
find ${OUTPUT_DIR}/node_modules -iname "*.html" -o -iname "*.rst" -type f -delete
mv ${OUTPUT_DIR}/index.html ${OUTPUT_DIR}/api-docs.html

echo "+++ Reference docs created"
