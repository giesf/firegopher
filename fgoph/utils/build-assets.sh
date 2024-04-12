#!/bin/bash
source ./config.sh

mkdir -p $ASSET_DIR

./build-init.sh
./download-kernel.sh
./rebuild.sh