#!/bin/sh
# Copyright Jetstack Ltd. See LICENSE for details.

set -o errexit
set -o nounset
set -o pipefail

# configure elrepo
rpm --import https://www.elrepo.org/RPM-GPG-KEY-elrepo.org
rpm -ivh http://www.elrepo.org/elrepo-release-7.0-2.el7.elrepo.noarch.rpm

# enable repo and install latest kernel
yum --enablerepo=elrepo-kernel install -y kernel-ml
