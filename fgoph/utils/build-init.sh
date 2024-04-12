#!/bin/bash
source ./config.sh
echo "Changing directory to init project..."
cd ../guest-init

echo "Building init project..."
set -e
export CGO_ENABLED=0
go build -o $ASSET_DIR/bin/guest-init
ldd $ASSET_DIR/bin/guest-init
cd ../utils

