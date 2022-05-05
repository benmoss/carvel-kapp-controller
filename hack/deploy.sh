#!/bin/bash

set -e

docker build -t ko.local/kc-base -q .
export VERSION=develop
ytt -f config | ko resolve -Pf- | kapp deploy -a kc -f- -c -y

source ./hack/secretgen-controller.sh
deploy_secretgen-controller
