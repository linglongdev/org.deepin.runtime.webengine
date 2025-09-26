#!/bin/bash
set -x
set -e
VERSION="$1"
PROJECT="$PWD"

# 获取base的包列表，用于跳过重复包
ll-builder build --skip-pull-depend --skip-fetch-source --skip-commit-output -- cp /packages.list ./base_packages.list

echo "Package: qemu-user-static" >> base_packages.list

rm -rf vscode-linglong || true
git clone https://github.com/myml/vscode-linglong.git --depth 1
cd vscode-linglong


for file in linglong.yaml arm64/linglong.yaml loong64/linglong.yaml sw64/linglong.yaml mips64/linglong.yaml; do
    sed -i "s#  version: .*#  version: ${VERSION}#" ../$file
    go run ./src/tools ../$file $PROJECT/base_packages.list
done

rm $PROJECT/base_packages.list