#!/bin/sh
# Copyright Jetstack Ltd. See LICENSE for details.

set -o errexit
set -o nounset
set -o pipefail

mkdir -p /tmp/packer-puppet-masterless
chown centos:centos /tmp/packer-puppet-masterless
