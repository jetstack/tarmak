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

set -e

REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY:-"$(cd "$(dirname "$0")" && pwd -P)"/..}
cd "${REPO_ROOT}"

GIT_TAG=$(git describe --tags --abbrev=0)

REFERENCE_PATH="docs/generated/reference"
REFERENCE_ROOT=$(cd "${REPO_ROOT}/${REFERENCE_PATH}" 2> /dev/null && pwd -P)
OUTPUT_DIR="${REFERENCE_ROOT}/includes"

BINDIR=$REPO_ROOT/bin
HACKDIR=$REPO_ROOT/hack

echo "+++ Removing old output"
rm -Rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

echo "+++ Running openapi-gen"
${BINDIR}/openapi-gen \
        --input-dirs github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1,k8s.io/apimachinery/pkg/version \
        --output-package "github.com/jetstack/tarmak/${REFERENCE_ROOT}/openapi"
        #github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1,\
        #github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1, \

## Generate swagger.json from the Golang generated openapi spec
echo "+++ Running 'swagger-gen' to generate swagger.json"
mkdir -p "${REFERENCE_ROOT}/openapi-spec"
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
IMAGE_NAME="tarmak-update-reference-docs:${GIT_TAG}"
BRODOC_DIR=${HACKDIR}/brodocs
rm -rf ${OUTPUT_DIR}
mkdir -p ${OUTPUT_DIR}
cp -r ${BRODOC_DIR} ${REFERENCE_ROOT}/brodocs
docker build -t ${IMAGE_NAME} ${REFERENCE_ROOT}
CONTAINER_ID=$(docker create ${IMAGE_NAME})
docker start -a ${CONTAINER_ID}
docker cp ${CONTAINER_ID}:/docs/index.html ${OUTPUT_DIR}/.
docker cp ${CONTAINER_ID}:/docs/navData.js ${OUTPUT_DIR}/.
docker rm ${CONTAINER_ID}
cp -r ${BRODOC_DIR}/{node_modules,*.js,*.css} ${OUTPUT_DIR}/.
rm -rf ${REFERENCE_ROOT}/brodocs
rm -rf ${REFERENCE_ROOT}/includes
rm -rf ${REFERENCE_ROOT}/static_includes

echo "+++ Reference docs created"
