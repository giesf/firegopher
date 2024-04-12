#!/bin/bash
# Copyright 2023 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# fail if we encounter an error, uninitialized variable or a pipe breaks
source ./config.sh
set -eu -o pipefail

set -x
PS4='+\t '

cd $(dirname $0)
ARCH=$(uname -m)
OUTPUT_DIR=$ASSET_DIR

mkdir -p $OUTPUT_DIR
# Make sure we have all the needed tools
function install_dependencies {
    sudo apt update
    sudo apt install -y bc flex bison gcc make libelf-dev libssl-dev squashfs-tools busybox-static tree cpio curl
}

function dir2ext4img {
    # ext4
    # https://unix.stackexchange.com/questions/503211/how-can-an-image-file-be-created-for-a-directory
    local DIR=$1
    local IMG=$2
    # Default size for the resulting rootfs image is 300M
    local SIZE=${3:-300M}
    local TMP_MNT=$(mktemp -d)
    truncate -s "$SIZE" "$IMG"
    mkfs.ext4 -F "$IMG"
    sudo mount "$IMG" "$TMP_MNT"
    sudo tar c -C $DIR . |sudo tar x -C "$TMP_MNT"
    # cleanup
    sudo umount "$TMP_MNT"
    rmdir $TMP_MNT
}


# Build a rootfs
function build_rootfs {
    local ROOTFS_NAME=$1
    local flavour=${2}
    local FROM_CTR=public.ecr.aws/ubuntu/ubuntu:$flavour
    local rootfs="tmp_rootfs"
    mkdir -pv "$rootfs" "$OUTPUT_DIR"

    cp -rvf overlay/* $rootfs

    # curl -O https://cloud-images.ubuntu.com/minimal/releases/jammy/release/ubuntu-22.04-minimal-cloudimg-amd64-root.tar.xz
    #
    # TBD use systemd-nspawn instead of Docker
    #   sudo tar xaf ubuntu-22.04-minimal-cloudimg-amd64-root.tar.xz -C $rootfs
    #   sudo systemd-nspawn --resolv-conf=bind-uplink -D $rootfs

    # @TODO use podman instead of docker
    docker run --env rootfs=$rootfs --privileged --rm -i -v "$PWD:/work" -w /work "$FROM_CTR" bash -s <<'EOF'

./chroot.sh


apt-get update
apt-get install -y ca-certificates 
apt-get install -y curl
apt-get install -y python3
rm -rf /var/lib/apt/lists/*


# Copy everything we need to the bind-mounted rootfs image file
dirs="bin etc home lib lib64 root sbin usr"
for d in $dirs; do tar c "/$d" | tar x -C $rootfs; done

# Make mountpoints
mkdir -pv $rootfs/{dev,proc,sys,run,tmp,var/lib/systemd}
EOF

    #sudo useradd -M www
    echo |sudo tee $rootfs/etc/resolv.conf

    # # Generate key for ssh access from host
    # if [ ! -s id_rsa ]; then
    #     ssh-keygen -f id_rsa -N ""
    # fi
    # sudo install -d -m 0600 "$rootfs/root/.ssh/"
    # sudo cp id_rsa.pub "$rootfs/root/.ssh/authorized_keys"
    # id_rsa=$OUTPUT_DIR/$ROOTFS_NAME.id_rsa
    # sudo cp id_rsa $id_rsa

    # -comp zstd but guest kernel does not support
    #rootfs_img="$OUTPUT_DIR/$ROOTFS_NAME.squashfs"
   # sudo mv $rootfs/root/manifest $OUTPUT_DIR/$ROOTFS_NAME.manifest
    #sudo mksquashfs $rootfs $rootfs_img -all-root -noappend
    rootfs_ext4=$OUTPUT_DIR/ubuntu.ext4
    dir2ext4img $rootfs $rootfs_ext4
    sudo rm -rf $rootfs
    sudo chown -Rc $USER. $OUTPUT_DIR
}


#### main ####

# install_dependencies

build_rootfs ubuntu-22.04 jammy


tree -h $OUTPUT_DIR