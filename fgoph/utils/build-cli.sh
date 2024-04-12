#!/bin/bash
#source ./config.sh
echo "Changing directory to cli project..."
cd ../cli

echo "Building cli project..."
set -e
#export CGO_ENABLED=0
go build -o /usr/bin/fgoph
ldd /usr/bin/fgoph
cd ../utils

