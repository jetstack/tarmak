#!/bin/bash

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

# This script should be run via `bazel run //hack:update-reference-docs`
REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY:-"$(cd "$(dirname "$0")" && pwd -P)"/..}
runfiles="${runfiles:-$(pwd)}"
export PATH="${runfiles}/hack/bin:${runfiles}/hack/brodocs:${PATH}"
cd "${REPO_ROOT}"

REFERENCE_PATH="docs/generated/reference"
REFERENCE_ROOT=$(cd "${REPO_ROOT}/${REFERENCE_PATH}" 2> /dev/null && pwd -P)
OUTPUT_DIR="${REFERENCE_ROOT}/includes"

BINDIR=$REPO_ROOT/bin

## cleanup removes files that are leftover from running various tools and not required
## for the actual output
#cleanup() {
#    pushd "${REFERENCE_ROOT}"
#    echo "+++ Cleaning up temporary docsgen files"
#    # Clean up old temporary files
#    rm -Rf "openapi-spec" "includes" "manifest.json"
#    popd
#}

# Ensure we start with a clean set of directories
#trap cleanup EXIT
#cleanup
echo "+++ Removing old output"
rm -Rf "${OUTPUT_DIR}"

#echo "+++ Creating temporary output directories"

## Generate swagger.json from the Golang generated openapi spec
#echo "+++ Running 'swagger-gen' to generate swagger.json"
#mkdir -p "${REFERENCE_ROOT}/openapi-spec"
## Generate swagger.json
## TODO: can we output to a tmpfile instead of in the repo?
#swagger-gen > "${REFERENCE_ROOT}/openapi-spec/swagger.json"

echo "+++ Running gen-apidocs"
# Generate Markdown docs
=======
cleanup() {
    pushd "${REFERENCE_ROOT}"
    echo "+++ Cleaning up temporary docsgen files"
    # Clean up old temporary files
    rm -Rf "openapi-spec" "includes" "manifest.json"
    popd
}

# Ensure we start with a clean set of directories
#trap cleanup EXIT
cleanup
echo "+++ Removing old output"
rm -Rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

echo "+++ Running openapi-gen"
${BINDIR}/openapi-gen \
        --input-dirs github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1,k8s.io/apimachinery/pkg/version \
        --output-package "github.com/jetstack/tarmak/${REFERENCE_PATH}/openapi"
        #github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1,\
        #github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1,\

## Generate swagger.json from the Golang generated openapi spec
echo "+++ Running 'swagger-gen' to generate swagger.json"
mkdir -p "${REFERENCE_ROOT}/openapi-spec"
go build -o ${REFERENCE_ROOT}/swagger-gen/swagger-gen ${REFERENCE_ROOT}/swagger-gen/main.go
${REFERENCE_ROOT}/swagger-gen/swagger-gen

echo "+++ Running gen-apidocs"

$BINDIR/gen-apidocs \
    --copyright "<a href=\"https://jetstack.io\">Copyright 2018 Jetstack Ltd.</a>" \
    --title "Tarmak API Reference" \
    --config-dir "${REFERENCE_ROOT}"

echo "+++ Running brodocs"
mkdir -p "${OUTPUT_DIR}"
INCLUDES_DIR="${REFERENCE_ROOT}/includes" \
OUTPUT_DIR="${OUTPUT_DIR}" \
MANIFEST_PATH="${REFERENCE_ROOT}/manifest.json" \
node ./hack/brodocs/brodoc.js ${MANIFEST_PATH} ${INCLUDES_DIR} ${OUTPUT_DIR}
