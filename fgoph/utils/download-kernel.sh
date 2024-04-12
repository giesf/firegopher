#!/bin/bash
source ./config.sh
echo ":$ASSET_DIR"
echo "🐧️ Downloading Linux kernel binary to $ASSET_DIR/kernel..."
ARCH="$(uname -m)"
wget -O $ASSET_DIR/kernel https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v1.5/${ARCH}/vmlinux-5.10.186
echo "✅ Kernel binary has downloaded and been written to $ASSET_DIR/kernel"
