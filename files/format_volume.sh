#!/bin/bash

if [ "$(id -u)" != "0" ]; then
  echo "Sorry, you are not root."
  exit 1
fi

/usr/sbin/blkid /dev/xvdd || (/usr/sbin/wipefs -fa /dev/xvdd && /usr/sbin/mkfs.xfs /dev/xvdd && mount /dev/xvdd /mnt && chown etcd:etcd /mnt && chmod 0750 /mnt && umount /mnt)
