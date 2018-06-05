#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE}")/.."; pwd)

find "${REPO_ROOT}/dist/" -executable -type f -print0 | xargs -0 upx
