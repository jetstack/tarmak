#!/bin/bash

# Copyright 2019 The Jetstack cert-manager contributors.
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

# This script should be run via `bazel run //hack:update-deps`
REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY:-"$(cd "$(dirname "$0")" && pwd -P)"/..}
runfiles="$(pwd)"
export LANG=C
cd "${REPO_ROOT}"

echo "+++ Running dep ensure"
dep ensure -v "$@"
echo "+++ Cleaning up circullar symlinks"
find vendor/ -follow -printf "" 2>&1 | grep "loop detected" | awk '{split($0, a, "‘|’"); print a[2];}' | xargs rm -f || true

echo "+++ Deleting bazel related data in vendor/"
find vendor/ -type f \( -name BUILD -o -name BUILD.bazel -o -name WORKSPACE \) -delete

touch vendor/BUILD.bazel
hack/update-bazel.sh
