#!/bin/bash
set -eu

install_dir=/fgoph/releases

bin_dir=/usr/bin
release_url="https://firegopher.dev"
latest='0.0.1'

mkdir -p "${install_dir}"
if [ -d "${install_dir}/${latest}" ]; then
        echo "${latest} already installed"
else
        mkdir -p "${install_dir}/${latest}"
        echo "downloading fghoph-v${latest}.zip to ${install_dir}"
        curl -o "${install_dir}/${latest}/fgoph-v${latest}.zip" -L "${release_url}/fgoph-v${latest}.zip"
        cd "${install_dir}/${latest}/"

        echo "decompressing fghoph-v${latest}.zip in ${install_dir}"
        unzip "fgoph-v${latest}.zip"
        rm "fgoph-v${latest}.zip"

        echo "linking fgoph"
        sudo ln -sfn "${install_dir}/${latest}/fgoph-v${latest}" "${bin_dir}/fgoph"

        echo "fgoph ${latest}: ready"
        fgoph --help
fi