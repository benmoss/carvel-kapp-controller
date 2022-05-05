#!/bin/bash

set -e -x -u

mkdir -p tmp/

# makes the get_kappctrl_ver function available (scrapes version from git tag)
source $(dirname "$0")/version-util.sh

docker build -t ko.local/kc-base -q --pull .
export VERSION="$(get_kappctrl_ver)"
ytt -f config/ -v kapp_controller_version="$VERSION" | ko resolve -Pf- > ./tmp/release.yml

shasum -a 256 ./tmp/release*.yml | tee ./tmp/checksums.txt

echo SUCCESS
