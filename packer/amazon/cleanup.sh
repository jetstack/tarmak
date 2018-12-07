#!/bin/sh
# Copyright Jetstack Ltd. See LICENSE for details.

set -o errexit
set -o nounset
set -o pipefail

# cleanup existing network configs
rm -f /etc/sysconfig/network-scripts/ifcfg-ens*

# clearing out ssh host keys
rm -f /etc/ssh/ssh_host_*
