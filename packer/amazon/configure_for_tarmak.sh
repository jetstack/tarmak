#!/bin/sh
# Copyright Jetstack Ltd. See LICENSE for details.

set -o errexit
set -o nounset
set -o pipefail

# install puppet repositories
rpm -ivh https://yum.puppetlabs.com/puppetlabs-release-pc1-el-7.noarch.rpm

# update all packages
yum update -y

# enable epel release
yum install -y epel-release
yum install -y git puppet-agent vim tmux socat python-pip at jq unzip awscli

# ensure aws cli works
aws help 2> /dev/null > /dev/null ||  { yum remove -y awscli && pip install awscli==1.16.68; }

# setup kernel parameters
sed -i '/GRUB_CMDLINE_LINUX=/c\\GRUB_CMDLINE_LINUX=\"console=tty0 crashkernel=0 console=ttyS0,115200 biosdevname=0 net.ifnames=0\"' /etc/sysconfig/grub
sed -i '/GRUB_CMDLINE_LINUX=/c\\GRUB_CMDLINE_LINUX=\"console=tty0 crashkernel=0 console=ttyS0,115200 biosdevname=0 net.ifnames=0\"' /etc/default/grub
grub2-mkconfig -o /boot/grub2/grub.cfg

# disable kdump service
systemctl disable kdump.service
