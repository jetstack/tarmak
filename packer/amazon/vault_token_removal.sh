#!/bin/sh
# Copyright Jetstack Ltd. See LICENSE for details.

set -o errexit
set -o nounset
set -o pipefail

rm -f /etc/vault/init-token
