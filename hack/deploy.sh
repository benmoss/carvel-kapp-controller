#!/bin/bash

set -e

docker build -t ko.local/kc-base -q .
export VERSION=develop
ytt -f config -v kapp_controller_version="$VERSION" | ko resolve -Pf- | kapp deploy -a kc -f- -c -y

source ./hack/secretgen-controller.sh
deploy_secretgen-controller
