#!/bin/bash

set -u -e

exit_with_error() {
  echo $1
  exit 1
}

GLOBAL_IPAM_BIN_SRC=/global-ipam
GLOBAL_IPAM_DST=/opt/cni/bin/global-ipam
GLOBAL_IPAM_RERWITE_HOST_LOCAL_DST=/opt/cni/bin/host-local
GLOBAL_IPAM_RERWITE_HOST_LOCAL_DST_OLD=/opt/cni/bin/host-local.old

yes | cp -f $GLOBAL_IPAM_BIN_SRC $GLOBAL_IPAM_DST || exit_with_error "Failed to copy $GLOBAL_IPAM_BIN_SRC to $GLOBAL_IPAM_DST"
yes | cp -f $GLOBAL_IPAM_RERWITE_HOST_LOCAL_DST $GLOBAL_IPAM_RERWITE_HOST_LOCAL_DST_OLD || exit_with_error "Failed to backup $GLOBAL_IPAM_RERWITE_HOST_LOCAL_DST to $GLOBAL_IPAM_RERWITE_HOST_LOCAL_DST_OLD"
yes | cp -f $GLOBAL_IPAM_BIN_SRC $GLOBAL_IPAM_RERWITE_HOST_LOCAL_DST || exit_with_error "Failed to write $GLOBAL_IPAM_BIN_SRC to $GLOBAL_IPAM_RERWITE_HOST_LOCAL_DST"

echo "install global-ipam binary"

./cni-server
